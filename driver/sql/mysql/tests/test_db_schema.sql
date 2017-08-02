
create database if not exists seedr_test;

use seedr_test;

drop table if exists users;
create table users (
    id int(10) unsigned not null auto_increment,
    email      varchar(250) not null,
    name       varchar(250) not null,
    active     boolean default false,
    checkin    datetime,
    created_at datetime not null default NOW(),
    primary key (id)
) engine=InnoDB default charset=utf8;

drop table if exists articles;
create table articles (
    id           int(10) unsigned not null auto_increment,
    author_id    int(10) unsigned not null,
    title        varchar(250) not null,
    body         text not null,
    created_at   datetime not null default NOW(),

    primary key (id)
) engine=InnoDB default charset=utf8;

drop table if exists clubs;
create table clubs (
    id          int(10) unsigned not null auto_increment,
    name        varchar(250) not null,

    primary key (id)
) engine=InnoDB default charset=utf8;

drop table if exists clubs_to_users;
create table clubs_to_users (
    id          int(10) unsigned not null auto_increment,
    club_id     int(10) unsigned not null,
    user_id     int(10) unsigned not null,

    primary key (id)
) engine=InnoDB default charset=utf8;


drop table if exists hellota_fields;
create table hellota_fields (
    id int(10) unsigned not null auto_increment,
    a int(10) unsigned,
    b text,
    c varchar(200),
    d date,
    e boolean,
    f int(10) unsigned,
    g text,
    h varchar(200),
    i date,
    j boolean,
    k int(10) unsigned,
    l text,
    m varchar(200),
    n date,
    o boolean,
    p int(10) unsigned,
    q text,
    r varchar(200),
    s date,
    t boolean,
    u int(10) unsigned,
    v text,
    w varchar(200),
    x date,
    y boolean,
    z int(10) unsigned,
    primary key (id)
) engine=InnoDB default charset=utf8;
