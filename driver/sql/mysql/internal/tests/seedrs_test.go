package tests

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"testing"

	"github.com/go-sql-driver/mysql"
	"github.com/josephbuchma/seedr"
	"github.com/josephbuchma/seedr/driver/sql/mysql/internal/tests/models"
	"github.com/josephbuchma/seedr/driver/sql/mysql/internal/tests/seedrs"
	"github.com/josephbuchma/seedr/driver/sql/mysql/internal/tests/util"
)

var schemaSQL = flag.String("schema", "test_db_schema.sql", "test db schema")

func cleanDB() {
	wd, wderr := os.Getwd()
	fmt.Println(wd, wderr)

	cmd := exec.Command("bash", "-c", fmt.Sprintf("mysql -h 127.0.0.1 -u root < %s", *schemaSQL))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stdout
	err := cmd.Run()
	if err != nil {
		panic(err)
	}
}

func TestMain(m *testing.M) {
	flag.Parse()
	os.Exit(m.Run())
}

var sdr = seedrs.TestSeedr

func TestSeedrs(t *testing.T) {
	cleanDB()

	t.Run("Basic insert", func(t *testing.T) {
		expextedUser := models.User{
			ID:        1,
			Name:      "Agent Smith 1",
			Email:     "agentsmith-1@gmail.com",
			Active:    false,
			Checkin:   mysql.NullTime{Time: util.TestTime, Valid: true},
			CreatedAt: util.TestTime,
		}
		u := models.User{}
		sdr.Create("InactiveUser").Scan(&u)
		util.AssertDeepEqual(t, expextedUser, u)
	})

	t.Run("Insert with ForeignKey", func(t *testing.T) {
		expectedArticle := models.Article{
			ID:        1,
			UserID:    2,
			Title:     "Awesome Title 1",
			Body:      "test body value",
			CreatedAt: util.TestTime,
		}

		expectedAuthor := models.User{
			ID:        2,
			Name:      "Agent Smith 2",
			Email:     "agentsmith-2@gmail.com",
			Active:    true,
			Checkin:   mysql.NullTime{Time: util.TestTime, Valid: true},
			CreatedAt: util.TestTime,
		}

		a := models.Article{}
		u := models.User{}
		ins := sdr.Create("TestArticle")
		ins.Scan(&a)
		ins.ScanRelated("author", &u)

		util.AssertDeepEqual(t, expectedArticle, a)
		util.AssertDeepEqual(t, expectedAuthor, u)
	})

	t.Run("Batch insert with ForeignKey", func(t *testing.T) {
		expectedArticles := []models.Article{
			{
				ID:        2,
				UserID:    3,
				Title:     "Awesome Title 2",
				Body:      "test body value",
				CreatedAt: util.TestTime,
			},
			{
				ID:        3,
				UserID:    4,
				Title:     "Awesome Title 3",
				Body:      "test body value",
				CreatedAt: util.TestTime,
			},
		}

		expectedAuthors := []models.User{
			{
				ID:        3,
				Name:      "Agent Smith 3",
				Email:     "agentsmith-3@gmail.com",
				Active:    true,
				Checkin:   mysql.NullTime{Time: util.TestTime, Valid: true},
				CreatedAt: util.TestTime,
			},
			{
				ID:        4,
				Name:      "Agent Smith 4",
				Email:     "agentsmith-4@gmail.com",
				Active:    true,
				Checkin:   mysql.NullTime{Time: util.TestTime, Valid: true},
				CreatedAt: util.TestTime,
			},
		}

		var articles []models.Article
		insArticles := sdr.CreateBatch("TestArticle", 2)
		insArticles.Scan(&articles)

		util.AssertDeepEqual(t, expectedArticles, articles)

		for i := 0; i < 2; i++ {
			u := models.User{}
			a := models.Article{}
			insArticles.Index(i).Scan(&a).ScanRelated("author", &u)

			util.AssertDeepEqual(t, expectedArticles[i], a)
			util.AssertDeepEqual(t, expectedAuthors[i], u)
		}
	})

	t.Run("Insert with batch child related records", func(t *testing.T) {
		usr := models.User{}
		articles := []models.Article(nil)
		sdr.Create("UserHeavyWriter").Scan(&usr).Related("articles").Scan(&articles)

		expextedUser := models.User{
			ID:        5,
			Name:      "Agent Smith 5",
			Email:     "agentsmith-5@gmail.com",
			Active:    true,
			Checkin:   mysql.NullTime{Time: util.TestTime, Valid: true},
			CreatedAt: util.TestTime,
		}
		expectedArticles := []models.Article{
			{
				ID:        4,
				UserID:    5,
				Title:     "Awesome Title 4",
				Body:      "test body value",
				CreatedAt: util.TestTime,
			},
			{
				ID:        5,
				UserID:    5,
				Title:     "Awesome Title 5",
				Body:      "test body value",
				CreatedAt: util.TestTime,
			},
		}

		util.AssertDeepEqual(t, expextedUser, usr)
		util.AssertDeepEqual(t, expectedArticles, articles)
	})

	t.Run("Batch insert with batch child related records", func(t *testing.T) {
		var usrs []models.User
		articles := []models.Article(nil)
		sdr.CreateBatch("UserHeavyWriter", 2).Scan(&usrs).Index(0).Related("articles").Scan(&articles)

		expextedUsers := []models.User{
			{
				ID:        6,
				Name:      "Agent Smith 6",
				Email:     "agentsmith-6@gmail.com",
				Active:    true,
				Checkin:   mysql.NullTime{Time: util.TestTime, Valid: true},
				CreatedAt: util.TestTime,
			},
			{
				ID:        7,
				Name:      "Agent Smith 7",
				Email:     "agentsmith-7@gmail.com",
				Active:    true,
				Checkin:   mysql.NullTime{Time: util.TestTime, Valid: true},
				CreatedAt: util.TestTime,
			},
		}
		util.AssertDeepEqual(t, expextedUsers, usrs)
	})

	t.Run("Many to many insert", func(t *testing.T) {
		usrs := []models.User(nil)
		var club models.Club
		cwu := sdr.Create("ClubWithUsers")
		cwu.Scan(&club)
		chusrs := cwu.Related("users")
		chusrs.Scan(&usrs)

		expectedClub := models.Club{
			ID:   1,
			Name: "Club-1",
		}

		util.AssertDeepEqual(t, expectedClub, club)

		expextedUsers := []models.User{
			{
				ID:        8,
				Name:      "Agent Smith 8",
				Email:     "agentsmith-8@gmail.com",
				Active:    true,
				Checkin:   mysql.NullTime{Time: util.TestTime, Valid: true},
				CreatedAt: util.TestTime,
			},
			{
				ID:        9,
				Name:      "Agent Smith 9",
				Email:     "agentsmith-9@gmail.com",
				Active:    true,
				Checkin:   mysql.NullTime{Time: util.TestTime, Valid: true},
				CreatedAt: util.TestTime,
			},
		}
		util.AssertDeepEqual(t, expextedUsers, usrs)
	})

	t.Run("CreateRelatedBatch M2M", func(t *testing.T) {
		club := models.Club{}
		usrs := []models.User{}
		sdr.Create("Club").CreateRelatedBatch("users", "TestUser", 2).Scan(&club).ScanRelated("users", &usrs)

		expextedUsers := []models.User{
			{
				ID:        10,
				Name:      "Agent Smith 10",
				Email:     "agentsmith-10@gmail.com",
				Active:    true,
				Checkin:   mysql.NullTime{Time: util.TestTime, Valid: true},
				CreatedAt: util.TestTime,
			},
			{
				ID:        11,
				Name:      "Agent Smith 11",
				Email:     "agentsmith-11@gmail.com",
				Active:    true,
				Checkin:   mysql.NullTime{Time: util.TestTime, Valid: true},
				CreatedAt: util.TestTime,
			},
		}
		util.AssertDeepEqual(t, models.Club{ID: 2, Name: "Club-2"}, club)
		util.AssertDeepEqual(t, expextedUsers, usrs)
	})

	t.Run("CreateRelatedBatch direct childs", func(t *testing.T) {
		usr := models.User{}
		articles := []models.Article{}
		sdr.Create("User").CreateRelatedBatch("articles", "Article", 2).Scan(&usr).ScanRelated("articles", &articles)

		expectedArticles := []models.Article{
			{
				ID:        10,
				UserID:    12,
				Title:     "Awesome Title 10",
				Body:      "test body value",
				CreatedAt: util.TestTime,
			},
			{
				ID:        11,
				UserID:    12,
				Title:     "Awesome Title 11",
				Body:      "test body value",
				CreatedAt: util.TestTime,
			},
		}

		util.AssertDeepEqual(t, models.User{
			ID:        12,
			Name:      "Agent Smith 12",
			Email:     "agentsmith-12@gmail.com",
			Active:    true,
			Checkin:   mysql.NullTime{Time: util.TestTime, Valid: true},
			CreatedAt: util.TestTime,
		}, usr)
		util.AssertDeepEqual(t, expectedArticles, articles)
	})

	t.Run("CreateCustom with M2M relation", func(t *testing.T) {
		usr := models.User{}
		club := models.Club{}
		sdr.CreateCustom("Club", seedr.Trait{
			"users": seedr.CreateRelated("User"),
		}).Scan(&club).ScanRelated("users", &usr)

		util.AssertDeepEqual(t, club, models.Club{ID: 3, Name: "Club-3"})
		util.AssertDeepEqual(t, usr, models.User{
			ID:        13,
			Email:     "agentsmith-13@gmail.com",
			Name:      "Agent Smith 13",
			Active:    true,
			Checkin:   mysql.NullTime{Time: util.TestTime, Valid: true},
			CreatedAt: util.TestTime,
		})
	})

	t.Run("CreateCustom with M2M relation with inline CreateRelatedCustom with child relation", func(t *testing.T) {
		usr := models.User{}
		club := models.Club{}
		article := models.Article{}
		sdr.CreateCustom("Club", seedr.Trait{
			"users": seedr.CreateRelatedCustom("User", seedr.Trait{
				"articles": seedr.CreateRelated("Article"),
			}),
		}).Scan(&club).
			ScanRelated("users", &usr).
			Related("users").
			ScanRelated("articles", &article)

		util.AssertDeepEqual(t, club, models.Club{ID: 4, Name: "Club-4"})
		util.AssertDeepEqual(t, usr, models.User{
			ID:        14,
			Email:     "agentsmith-14@gmail.com",
			Name:      "Agent Smith 14",
			Active:    true,
			Checkin:   mysql.NullTime{Time: util.TestTime, Valid: true},
			CreatedAt: util.TestTime,
		})
		util.AssertDeepEqual(t, article, models.Article{
			ID:        12,
			UserID:    14,
			Title:     "Awesome Title 12",
			Body:      "test body value",
			CreatedAt: util.TestTime,
		})
	})

	t.Run("Build instance", func(t *testing.T) {
		usr := models.User{}
		sdr.Build("TestUser").Scan(&usr)
		util.AssertDeepEqual(t, models.User{
			ID:        0,
			Name:      "Agent Smith 15",
			Email:     "agentsmith-15@gmail.com",
			Active:    true,
			Checkin:   mysql.NullTime{Time: util.TestTime, Valid: true},
			CreatedAt: util.TestTime,
		}, usr)
	})

}

const benchBatchSize = 10000

func BenchmarkInsertBatch(b *testing.B) {
	cleanDB()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		_ = sdr.CreateBatch("UserJohn", benchBatchSize)
	}
}

func BenchmarkInsertBatchWithRelations(b *testing.B) {
	cleanDB()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		_ = sdr.CreateBatch("TestArticle", benchBatchSize)
	}
}

func BenchmarkInsert(b *testing.B) {
	cleanDB()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		_ = sdr.Create("UserJohn")
	}
}

func BenchmarkInsertManyFields(b *testing.B) {
	cleanDB()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		_ = sdr.Create("HellotaFieldsTest")
	}
}
