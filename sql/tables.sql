drop table if exists page;
drop table if exists verification;
drop table if exists aut_book_interm;
drop table if exists book;
drop table if exists end_user;
drop table if exists administrator;
drop table if exists user_gen;
drop table if exists author;

create table user_gen (
	guid uuid NOT NULL DEFAULT uuid_generate_v4(),
	nickname varchar(50) NOT NULL,
	reg_date timestamptz NOT NULL DEFAULT now(),
	email varchar(50) NOT NULL,
	pass_hash varchar(64) NOT NULL,
	CONSTRAINT proper_email CHECK (email ~* '^.+@.+$'),
	CONSTRAINT proper_nickname CHECK (nickname ~* '^[^@]+$')
);

create table end_user (
	banned bool NOT NULL DEFAULT false,
	primary key(guid),
	CONSTRAINT unique_nickname_user unique (nickname),
	CONSTRAINT unique_email_user unique (email),
	CONSTRAINT unique_id_user unique (guid)
) inherits (user_gen);

create table administrator (
	rights bit(2) NOT NULL,
    check_num int NOT NULL DEFAULT 0,
	primary key(guid),
	CONSTRAINT unique_nickname_admin unique (nickname),
	CONSTRAINT unique_email_admin unique (email),
	CONSTRAINT unique_id_admin unique (guid)
) inherits (user_gen);

create table book (
    guid uuid NOT NULL unique DEFAULT uuid_generate_v4(),
    user_id uuid NOT NULL,
    book_name varchar(50) NOT NULL,
    book_genre genre NOT NULL,
    write_year smallint DEFAULT 2022,
    page_num smallint DEFAULT 0,
    primary key(guid),
    foreign key (user_id) references end_user (guid)
);

create table page (
    guid uuid NOT NULL unique DEFAULT uuid_generate_v4(),
    book_id uuid NOT NULL,
    num smallint NOT NULL,
    page_data bytea NOT NULL,
    primary key(guid),
    foreign key (book_id) references book (guid) on delete cascade
);

create table verification (
    guid uuid NOT NULL unique DEFAULT uuid_generate_v4(),
    book_id uuid NOT NULL unique,
    admin_id uuid NOT NULL DEFAULT get_admin(),
    primary key(guid),
    foreign key (book_id) references book (guid) on delete cascade,
    foreign key (admin_id) references administrator (guid) on delete set default
);

create table author (
    guid uuid NOT NULL unique DEFAULT uuid_generate_v4(),
    author_name varchar(50) NOT NULL,
    CONSTRAINT unique_author unique (author_name),
    primary key(guid)
);

create table aut_book_interm (
    author_id uuid NOT NULL,
    book_id uuid NOT NULL,
    foreign key (book_id) references book (guid) on delete cascade,
    foreign key (author_id) references author (guid)
);