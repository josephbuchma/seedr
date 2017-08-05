// Package seedr is inspired by Ruby's factory_girl.
// It's a fixtures replacement that allows to
// define factories and build/store objects
// in expressive and flexible way.
package seedr

import (
	"database/sql"
	"fmt"
	"reflect"
	"regexp"
	"strings"
	"time"

	"github.com/josephbuchma/seedr/driver"
	"github.com/josephbuchma/seedr/driver/noop"
)

const (
	// Include is a special key for Trait definition
	// that defines on what traits current trait is based on.
	// Value must be a string with names of traits (of current factory only)
	// separated by whitespace.
	Include = "SEEDR_INCLUDE_TRAITS"
)

// Trait contains factory fields definitions.
// Key is a field name and value is any of supported:
//    - any numeric type
//    - time.Time
//    - sql.Scanner
//    - Generator
//    - nil (interface{}(nil))
// Special key is `Include` constant
// that defines on what traits this Trait is based on.
// Every field of trait is a name of respective field
// in database.
type Trait map[string]interface{}

var traitNameRegexp = regexp.MustCompile(`\w+`)

// includes returns list of traits to include
func (t Trait) includes() []string {
	if _, ok := t[Include]; !ok {
		return nil
	}
	return traitNameRegexp.FindAllString(t[Include].(string), -1)
}

// merge adds new keys from given trait
func (t Trait) merge(trait Trait, override bool) Trait {
	for field, val := range trait {
		if field == Include {
			continue
		}
		if override {
			t[field] = val
		} else if _, ok := t[field]; !ok {
			t[field] = val
		}
	}
	return t
}

// getFieldValue bypasses given value if it's of supported type
// or returns .Next() if it's a Generator
// otherwise it panics
func getFieldValue(v interface{}) (interface{}, error) {
	if v == nil {
		return nil, nil
	}
	switch v := v.(type) {
	case int, uint, int8, uint8, int16,
		uint16, int32, uint32, int64, uint64,
		float32, float64, string, []byte, bool,
		time.Time, sql.Scanner, auto, *relationField:
		return v, nil
	case Generator:
		return getFieldValue(v.Next())
	}
	return nil, fmt.Errorf("`%v` is of unsupported type %T", v, v)
}

// FactoryConfig ...
type FactoryConfig struct {
	factoryName string
	// Entity is a name of database schema/table, index, etc.
	// Entity may not be required, depends on driver.
	// If not given, factory name is used.
	Entity string
	// PrimaryKey is a name of primary key field.
	// It is required if you define any relations.
	// It also may be required by your Driver.
	PrimaryKey string
}

func (fc FactoryConfig) mustPK() string {
	if fc.PrimaryKey == "" {
		panicf("PrimaryKey for %s is not defined", fc.Entity)
	}
	return fc.PrimaryKey
}

// Factory is a container for factory definition
type Factory struct {
	FactoryConfig
	Relations
	Traits
}

// Traits is a map of Trait declarations (name -> Trait).
// Capitalized names are 'exported', others are private (e.g. only can be included by other Trait of this Traits).
type Traits map[string]Trait

// mergeIncludes recursively merge all includes of given trait.
// It panics on circular import.
func (t Traits) mergeIncludes(trait Trait, traitName string) Trait {
	ret := Trait{}
	for _, incl := range trait.includes() {
		if tr, ok := t[incl]; ok {
			if strings.Contains(toString(tr[Include]), traitName) {
				panicf("Circular include %q <-> %q", traitName, incl)
			}
			ret.merge(t.mergeIncludes(tr, incl), true)
		} else {
			panicf("Invalid include in %q: trait %q does not exist", traitName, incl)
		}
	}
	return ret.merge(trait, true)
}

// build merge all includes and returns list of public trait names
func (t Traits) buildPublics() []string {
	var publics []string
	for name, trait := range t {
		if strings.ToUpper(name[0:1]) == name[0:1] {
			t[name] = t.mergeIncludes(trait, name)
			publics = append(publics, name)
		}
	}
	return publics
}

// ConfigFunc is used to configure Seedr
type ConfigFunc func(*Seedr)

// SetCreateDriver sets driver for `Create*` methods
func SetCreateDriver(b driver.Driver) ConfigFunc {
	if b == nil {
		panic("driver can't be nil")
	}
	return func(s *Seedr) {
		s.createDriver = b
	}
}

// SetBuildDriver sets driver for `Build*` methods
func SetBuildDriver(b driver.Driver) ConfigFunc {
	if b == nil {
		panic("driver can't be nil")
	}
	return func(s *Seedr) {
		s.buildDriver = b
	}
}

// MapFieldFunc returns name of Trait's field based on StructField
type MapFieldFunc func(reflect.StructField) (traitFieldName string, err error)

// SnakeFieldMapper converts field name to snake case;
// acronyms are converted to lower-case and preceded by an underscore.
// Note that multiple consequent acronyms will be considered a single acronym
// (e.g. ACDCTime will become acdc_time, not ac_dc_time)
func SnakeFieldMapper() MapFieldFunc {
	return func(f reflect.StructField) (string, error) {
		return toSnake(f.Name), nil
	}
}

// NoopFieldMapper returns field's `Name` without changes
func NoopFieldMapper() MapFieldFunc {
	return func(f reflect.StructField) (string, error) {
		return f.Name, nil
	}
}

// TagFieldMapper looks for specific struct field tag or uses fallback mapper if tag is not found.
func TagFieldMapper(tag string, fallback MapFieldFunc) MapFieldFunc {
	return func(f reflect.StructField) (string, error) {
		v := f.Tag.Get(tag)
		if v != "" {
			return strings.Split(v, ",")[0], nil
		}
		if fallback != nil {
			return fallback(f)
		}
		return "", fmt.Errorf("TagFieldMapper failed on struct field %#v", f)
	}
}

// RegexpTagFieldMapper applies regex to whole struct tag string and uses first
// match as result. If it fails fallback mapper is used.
func RegexpTagFieldMapper(regex string, fallback MapFieldFunc) MapFieldFunc {
	cmpRegex := regexp.MustCompile(regex)
	return func(f reflect.StructField) (string, error) {
		name := cmpRegex.FindStringSubmatch(string(f.Tag))
		if len(name) != 0 {
			return name[1], nil
		}
		if fallback != nil {
			return fallback(f)
		}
		return "", fmt.Errorf("TagFieldMapper failed on struct field %#v", f)
	}
}

// SetFieldMapper sets MapFieldFunc func for Seedr
func SetFieldMapper(f MapFieldFunc) ConfigFunc {
	if f == nil {
		panic("name field mapper can't be nil")
	}
	return func(s *Seedr) {
		s.extractFieldName = f
	}
}

// Seedr is a collection of factories.
type Seedr struct {
	name string
	// createDriver is used in Create* methods
	createDriver driver.Driver
	// buildDriver is used in Build* methods
	buildDriver driver.Driver
	// publicTraits contains map with all traits that starts with capital letter
	publicTraits     map[string]*publicTrait
	extractFieldName MapFieldFunc
}

// New creates Seedr instance with NoopFieldMapper
// and NoopDriver for create and build methods by default.
func New(name string, config ...ConfigFunc) *Seedr {
	sdr := &Seedr{
		name:             validString(name, "Seedr name can't be empty string"),
		publicTraits:     make(map[string]*publicTrait),
		extractFieldName: NoopFieldMapper(),
		createDriver:     noop.NoopDriver{},
		buildDriver:      noop.NoopDriver{},
	}
	for _, cfg := range config {
		cfg(sdr)
	}
	return sdr
}

// Add adds new factory to this seedr
func (sdr *Seedr) Add(factoryName string, f Factory) *Seedr {
	f.FactoryConfig.factoryName = factoryName
	if f.FactoryConfig.Entity == "" {
		f.FactoryConfig.Entity = factoryName
	}
	relations := f.Relations.normalize()
	t := f.Traits

	for _, name := range t.buildPublics() {
		trait := t[name]
		if _, exist := sdr.publicTraits[name]; exist {
			panicf("Public trait with name %q already exists", name)
		}
		extRelCnt := 0
		for _, rel := range relations {
			for f := range trait {
				if rel.kind != relationParent && relations[f] != nil {
					extRelCnt++
				}
			}
		}
		pub := &publicTrait{
			sdr:                  sdr,
			factory:              &f,
			name:                 name,
			trait:                trait,
			relations:            relations,
			externalRelationsCnt: extRelCnt,
		}
		pub.normalizeRelations()
		sdr.publicTraits[name] = pub
	}
	return sdr
}

func (sdr *Seedr) getPublicTrait(name string) *publicTrait {
	t, ok := sdr.publicTraits[name]
	if !ok {
		panicf("Trait %q does not exist", name)
	}
	return t
}

type rawTrait struct {
	trait *publicTrait
	rels  map[string]*relationField
	data  []map[string]interface{}

	insertFields, returnFields []string

	iter int
}

func (it *rawTrait) addRel(field string, dep *relationField) {
	if it.rels == nil {
		it.rels = make(map[string]*relationField)
	}
	it.rels[field] = dep
}

type publicTrait struct {
	sdr                  *Seedr
	factory              *Factory
	name                 string
	trait                Trait
	relations            map[string]*Relation
	externalRelationsCnt int
}

func (t *publicTrait) normalizeRelations() {
	for relName, rel := range t.relations {
		if rel.kind == relationParent {
			if t.trait[relName] != nil {
				delete(t.trait, rel.lfield)
			}
		}
	}
}

func (t *publicTrait) next(n int, ovr Trait) *rawTrait {
	depsReady := false
	rt := &rawTrait{
		trait: t,
		data:  make([]map[string]interface{}, 0, n),
	}

	for i := 0; i < n; i++ {
		nxt := make(Trait)
		for i, trait := range []Trait{t.trait, ovr} {
			for k, v := range trait {
				if ovr != nil {
					if ov, ok := ovr[k]; ok {
						if i == 0 {
							v = ov
						} else if _, ok := t.trait[k]; ok {
							continue
						}
					}
				}
				fv, err := getFieldValue(v)
				if err != nil {
					panicf("Failed to get value of field %q: %s", k, err)
				}
				switch fv.(type) {
				case *relationField, auto:
				default:
					nxt[k] = fv
				}
				if depsReady {
					continue
				}
				switch v := fv.(type) {
				case *relationField:
					if v.lfield == "" {
						if rel, ok := t.relations[k]; ok {
							if _, ok := t.sdr.publicTraits[v.traitName]; !ok {
								panicf("Trait %q does not exist (trait %q, field %q)", v.traitName, t.name, k)
							}
							v.kind = rel.kind
							v.rfield = rel.rfield
							v.lfield = rel.lfield
						} else {
							panicf("Relation %s is not defined for trait %s", k, t.name)
						}
					}
					rt.addRel(k, v)
					if v.kind == relationParent {
						rt.insertFields = append(rt.insertFields, v.lfield)
					}
				case auto:
					rt.returnFields = append(rt.returnFields, k)
				default:
					rt.insertFields = append(rt.insertFields, k)
				}
			}
		}
		depsReady = true
		rt.data = append(rt.data, nxt)

	}
	rt.returnFields = append(rt.insertFields, rt.returnFields...)
	return rt
}

func (t *publicTrait) build(ovr Trait, n int) (ret *TraitInstances) {
	return t.drvCreate(ovr, n, t.sdr.buildDriver)
}

func (t *publicTrait) create(ovr Trait, n int) (ret *TraitInstances) {
	return t.drvCreate(ovr, n, t.sdr.createDriver)
}

func (t *publicTrait) drvCreate(ovr Trait, n int, drv driver.Driver) (ret *TraitInstances) {
	ret = &TraitInstances{sdr: t.sdr, trait: t, childs: make(map[string][]*TraitInstances)}
	rt := t.next(n, ovr)
	if rt.rels != nil {
		ret.parents = make(map[string]*TraitInstances)
		childs := make(map[string]*relationField)
		m2ms := make(map[string]*relationField)

		// handle child and many to many relations after this one is created
		defer func() {
			if len(childs) == 0 && len(m2ms) == 0 {
				return
			}
			pks := make([]interface{}, len(ret.data))
			pkName := t.factory.FactoryConfig.mustPK()
			for i, r := range ret.data {
				pks[i] = r[pkName]
			}
			for field, rel := range childs {
				ins := t.sdr.getPublicTrait(rel.traitName).create((Trait{
					rel.lfield: mnSliceSeq(pks, rel.n),
				}).merge(rel.override, false), len(rt.data)*rel.n)
				ret.childs[field] = ins.chop(n)
			}
			for field, rel := range m2ms {
				rels := t.sdr.getPublicTrait(rel.traitName).create(rel.override, len(rt.data)*rel.n)
				relpks := make([]interface{}, rels.Len())
				for i, r := range rels.data {
					relpks[i] = r[pkName]
				}
				joinTraitName := t.relations[field].joinTrait
				// TODO: relate join table
				t.sdr.getPublicTrait(joinTraitName).create(Trait{
					rel.lfield: mnSliceSeq(pks, rel.n),
					rel.rfield: mnSliceSeq(relpks, rel.n),
				}, len(rt.data)*rel.n)
				ret.childs[field] = rels.chop(n)
			}
		}()

		for field, rel := range rt.rels {
			switch rel.kind {
			case relationParent:
				ins := t.sdr.getPublicTrait(rel.traitName).create(rel.override, len(rt.data))
				ret.parents[field] = ins
			case relationChild:
				childs[field] = rel
			case relationM2M:
				m2ms[field] = rel
			}
		}
		if len(ret.parents) > 0 {
			for i, d := range rt.data {
				for relation, rel := range rt.rels {
					if rel.kind != relationParent {
						continue
					}
					p := ret.parents[relation]
					d[rel.lfield] = p.data[i][p.trait.factory.FactoryConfig.mustPK()]
				}
			}
		}
	}

	var err error
	ret.data, err = drv.Create(driver.Payload{
		Entity:       t.factory.FactoryConfig.Entity,
		PrimaryKey:   t.factory.FactoryConfig.PrimaryKey,
		InsertFields: rt.insertFields,
		ReturnFields: rt.returnFields,
		Data:         rt.data,
	})
	if err != nil {
		panicf("Seedr Driver error: %s", err)
	}
	return ret
}

// TraitInstance is a created trait instance
type TraitInstance struct {
	sdr   *Seedr
	insts *TraitInstances
	i     int
}

// createRelated creates related traits. Works for child and M2M relations.
func (ti TraitInstance) createRelated(relation, traitName string, n int, override Trait) TraitInstance {
	var ins *TraitInstances
	rel, ok := ti.insts.trait.relations[relation]
	if !ok {
		panicf("%q does not have relation %q", ti.insts.trait.factory.FactoryConfig.factoryName, relation)
	}
	switch rel.kind {
	case relationChild:
		ovr := Trait{
			rel.lfield: ti.insts.data[ti.i][ti.insts.trait.factory.FactoryConfig.mustPK()],
		}
		if override != nil {
			ovr = override.merge(ovr, true)
		}
		ins = ti.sdr.CreateCustomBatch(traitName, n, ovr)
	case relationM2M:
		if nm := ti.sdr.getPublicTrait(traitName).factory.FactoryConfig.factoryName; nm != rel.factory {
			panicf("Invalid M2M: expected factory %s, got %s", rel.factory, nm)
		}
		ins = ti.sdr.CreateCustomBatch(traitName, n, override)
		// TODO: bind 'parent' rel to both
		ti.sdr.CreateCustomBatch(rel.joinTrait, n, Trait{
			rel.lfield: ti.insts.data[0][ti.insts.trait.factory.FactoryConfig.mustPK()],
			rel.rfield: SequenceFunc(func(i int) interface{} {
				return ins.data[i][ins.trait.factory.FactoryConfig.mustPK()]
			}, 0),
		})
	case relationParent:
		panicf("TraitInstance#CreateRelated does not support 'parent' relations. Please use 'Seedr#CreateCustom' or define trait in factory")
	default:
		panic("UNREACHABLE")
	}
	if recs, ok := ti.insts.childs[relation]; ok && len(recs) > ti.i {
		recs[ti.i].append(ins)
	} else if !ok {
		ti.insts.childs[relation] = []*TraitInstances{ins}
	} else if len(recs) == ti.i {
		ti.insts.childs[relation] = append(ti.insts.childs[relation], ins)
	} else {
		panic("NOOP")
	}
	return ti
}

// CreateRelated creates single related instance. Works for FK (child only) and M2M.
// It returns original TraitInstance, not one that was created. Fetch created one using Related method.
func (ti TraitInstance) CreateRelated(relation, traitName string) TraitInstance {
	return ti.createRelated(relation, traitName, 1, nil)
}

// CreateRelatedCustom creates single related instance with additional changes. Works for FK (child only) and M2M
// It returns original TraitInstance, not one that was created. Fetch created one using Related method.
func (ti TraitInstance) CreateRelatedCustom(relation, traitName string, overrides Trait) TraitInstance {
	return ti.createRelated(relation, traitName, 1, overrides)
}

// CreateRelatedBatch creates n related instances. Works for FK (child only) and M2M
// It returns original TraitInstance, not one that was created. Fetch created one using Related method.
func (ti TraitInstance) CreateRelatedBatch(relation, traitName string, n int) TraitInstance {
	return ti.createRelated(relation, traitName, n, nil)
}

// CreateRelatedCustomBatch creates n related instances with additional changes. Works for FK (child only) and M2M
// It returns original TraitInstance, not one that was created. Fetch created one using Related method.
func (ti TraitInstance) CreateRelatedCustomBatch(relation, traitName string, n int, overrides Trait) TraitInstance {
	return ti.createRelated(relation, traitName, n, overrides)
}

// Scan initializes given struct instance `v` by TraitInstance's values.
// `v` must be a pointer to struct.
func (ti TraitInstance) Scan(v interface{}) TraitInstance {
	val := reflect.ValueOf(v).Elem()
	t := val.Type()
	rec := ti.insts.data[ti.i]
	scannedCnt := 0
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		fname, err := ti.sdr.extractFieldName(f)
		if err != nil {
			panicf("Failed to Scan: %s", err)
		}
		if iv, ok := rec[fname]; !ok {
			continue
		} else {
			if err := convertAssign(val.Field(i).Addr().Interface(), iv); err != nil {
				panic(err)
			}
			scannedCnt++
		}
	}
	if len(rec) < scannedCnt {
		panicf("Could not scan some values from this object: %#v", ti.insts.data[ti.i])
	}
	return ti
}

// ScanRelated scans related TraitInstance(s) into v and returns this TraitInstance.
func (ti TraitInstance) ScanRelated(relationName string, v interface{}) TraitInstance {
	ti.Related(relationName).Scan(v)
	return ti
}

// parent returns parent TraitInstance by given FIELD (FK) name.
func (ti TraitInstance) parent(field string) *TraitInstances {
	parents, ok := ti.insts.parents[field]
	if !ok {
		return nil
	}
	return parents.slice(ti.i, ti.i+1)
	//return &TraitInstance{sdr: ti.sdr, insts: parents, i: ti.i}
}

// child returns child TraitInstances by given relation name
func (ti TraitInstance) child(relationName string) *TraitInstances {
	childs, ok := ti.insts.childs[relationName]
	if !ok {
		return nil
	}
	return childs[ti.i]
}

// Related returns related TraitInstances (that was created by CreateRelated*)
func (ti TraitInstance) Related(relationName string) *TraitInstances {
	rels := ti.child(relationName)
	if rels == nil {
		rels = ti.parent(relationName)
	}
	if rels == nil {
		panicf("%q factory has no relation %q", ti.insts.trait.factory.factoryName, relationName)
	}
	return rels
}

// TraitInstances is a collection of created trait instances
type TraitInstances struct {
	sdr   *Seedr
	trait *publicTrait
	// recs is a list of raw inserted records
	data []map[string]interface{}
	// parents is field -> relation[i] for recs[i]
	parents map[string]*TraitInstances
	// childs is field -> list of childs where list[i] belogs to data[i]
	childs map[string][]*TraitInstances
}

func (ti *TraitInstances) append(r *TraitInstances) {
	for _, rec := range r.data {
		ti.data = append(ti.data, rec)
	}
	for fld, recs := range r.parents {
		ti.parents[fld].append(recs)
	}
	for fld, lst := range r.childs {
		for _, recs := range lst {
			ti.childs[fld] = append(ti.childs[fld], recs)
		}
	}
}

func (ti *TraitInstances) slice(b, e int) *TraitInstances {
	ret := &TraitInstances{
		sdr:   ti.sdr,
		trait: ti.trait,
		// recs is a list of raw inserted records
		data:    ti.data[b:e],
		parents: make(map[string]*TraitInstances),
	}
	for k, v := range ti.parents {
		ret.parents[k] = v.slice(b, e)
	}
	if len(ti.childs) > 0 {
		ret.childs = make(map[string][]*TraitInstances)
		for f, ch := range ti.childs {
			ret.childs[f] = ch[b:e]
		}
	}
	return ret
}

func (ti *TraitInstances) chop(n int) []*TraitInstances {
	if len(ti.data)%n != 0 {
		panic("Can't chop")
	}
	ret := make([]*TraitInstances, n)
	chunk := len(ti.data) / n
	for i := 0; i < n; i++ {
		ret[i] = ti.slice(i*chunk, i*chunk+chunk)
	}
	return ret
}

// Len returns total count of trait instances in this collection
func (ti *TraitInstances) Len() int {
	return len(ti.data)
}

// Index returns TraitInstance by index
func (ti *TraitInstances) Index(index int) TraitInstance {
	return TraitInstance{sdr: ti.sdr, i: index, insts: ti}
}

// ScanRelated is a shortcut for #Index(0).ScanRelated()
// It panics if there is not exactly one instance in this TraitInstances.
func (ti *TraitInstances) ScanRelated(relationName string, dest interface{}) {
	switch ti.Len() {
	case 0:
		panic("Nothing to Scan")
	case 1:
		ti.Index(0).ScanRelated(relationName, dest)
	default:
		panic("calling ScanRelated when there is more than one instance is not allowed")
	}
}

// Scan initializes given list of struct instances `dest` by TraitInstance's values.
// `v` must be a pointer to slice of structs.
// But, if there is only one instance, dest can be pointer to struct.
func (ti *TraitInstances) Scan(dest interface{}) (ret *TraitInstances) {
	if ti.Len() == 0 {
		panic("Nothing to Scan")
	}
	v := reflect.ValueOf(dest)
	t := v.Type()
	if t.Kind() != reflect.Ptr {
		panicf("Scan works only with pointers, %s was given.", t.Kind())
	}
	if ti.Len() == 1 {
		if t.Elem().Kind() == reflect.Struct {
			ti.Index(0).Scan(dest)
			return ti
		}
	}
	if t.Kind() != reflect.Ptr || t.Elem().Kind() != reflect.Slice {
		panicf("InsertedRecords#Scan argument must be pointer to slice")
	}

	t = t.Elem().Elem()
	sls := reflect.MakeSlice(reflect.SliceOf(t), len(ti.data), len(ti.data))
	v.Elem().Set(sls)

	var tPtr reflect.Type
	if t.Kind() == reflect.Ptr {
		tPtr = t
		t = t.Elem()
	}

	for i := 0; i < len(ti.data); i++ {
		vi := sls.Index(i)
		val := vi
		if tPtr != nil {
			val = reflect.Zero(t)
			vi.Set(val.Addr())
		}
		scannedCnt := 0
		for fid := 0; fid < t.NumField(); fid++ {
			f := t.Field(fid)
			fname, err := ti.sdr.extractFieldName(f)
			if err != nil {
				panicf("Failed to Scan: %s", err)
			}
			if iv, ok := ti.data[i][fname]; !ok {
				continue
			} else {
				if err := convertAssign(val.Field(fid).Addr().Interface(), iv); err != nil {
					panic(err)
				}
				scannedCnt++
			}
		}
		if len(ti.data[i]) < scannedCnt {
			panicf("Could not scan some values: %#v", ti.data[i])
		}
	}
	return ti
}

// CreateCustom overrides values of trait definition and creates resulting trait.
// It uses "create" driver (see SetCreateDriver)
func (sdr *Seedr) CreateCustom(traitName string, override Trait) TraitInstance {
	t := sdr.getPublicTrait(traitName)
	ins := t.create(override, 1)
	return ins.Index(0)
}

// CreateCustomBatch overrides values of trait definition
// and creates n instances of resulting trait.
// It uses "create" driver (see SetCreateDriver)
func (sdr *Seedr) CreateCustomBatch(traitName string, n int, override Trait) *TraitInstances {
	t := sdr.getPublicTrait(traitName)
	return t.create(override, n)
}

// CreateBatch creates n instances of trait
// It uses "create" driver (see SetCreateDriver)
func (sdr *Seedr) CreateBatch(traitName string, n int) *TraitInstances {
	t := sdr.getPublicTrait(traitName)
	return t.create(nil, n)
}

// Create creates an instance of trait
// It uses "create" driver (see SetCreateDriver)
func (sdr *Seedr) Create(traitName string) TraitInstance {
	t := sdr.getPublicTrait(traitName)
	ins := t.create(nil, 1)
	return ins.Index(0)
}

// BuildCustom builds trait with additional changes.
// It uses "build" driver (see SetBuildDriver)
func (sdr *Seedr) BuildCustom(traitName string, override Trait) TraitInstance {
	t := sdr.getPublicTrait(traitName)
	ins := t.build(override, 1)
	return ins.Index(0)
}

// BuildCustomBatch builds n trait instances with additional changes.
// It uses "build" driver (see SetBuildDriver)
func (sdr *Seedr) BuildCustomBatch(traitName string, n int, override Trait) *TraitInstances {
	t := sdr.getPublicTrait(traitName)
	return t.build(override, n)
}

// BuildBatch builds n trait instances.
// It uses "build" driver (see SetBuildDriver)
func (sdr *Seedr) BuildBatch(traitName string, n int) *TraitInstances {
	t := sdr.getPublicTrait(traitName)
	return t.build(nil, n)
}

// Build builds trait.
// It uses "build" driver (see SetBuildDriver)
func (sdr *Seedr) Build(traitName string) TraitInstance {
	t := sdr.getPublicTrait(traitName)
	ins := t.build(nil, 1)
	return ins.Index(0)
}
