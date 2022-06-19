drop index if exists admin_guid;

create index admin_guid on administrator(guid) include (nickname, email);

drop index if exists verification_adm;

create index verification_adm on verification(admin_id) include (book_id);

drop index if exists book_user_id;

create index book_user_id on book(user_id) include (guid, book_name);

drop index if exists user_nickname;

create index user_nickname on end_user(nickname) include (email, guid, pass_hash);

drop index if exists user_email;

create index user_email on end_user(email) include (guid, pass_hash);

drop index if exists user_guid;

create index user_guid on end_user(guid) include (banned);

drop index if exists trgm_author_name;

create index trgm_author_name on author using gin (author_name gin_trgm_ops);

drop index if exists trgm_book_name;

create index trgm_book_name on book using gin (book_name gin_trgm_ops);

drop index if exists admin_book_id;

create index admin_book_id on admin_book(admin_id, guid) include (book_name);

drop index if exists user_book_id;

create index user_book_id on user_book(user_id) include (nickname, book_name, book_genre, write_year, page_num, author_name);

drop index if exists user_book_guid;

create index user_book_guid on user_book(guid) include (nickname, book_name, book_genre, write_year, page_num, author_name);

drop index if exists user_book_genre_year;

create index user_book_genre_year on user_book(book_genre, write_year) include (nickname, book_name, page_num, author_name);

drop index if exists trgm_user_book_author_name;

create index trgm_user_book_author_name on user_book using gin (author_name gin_trgm_ops);

drop index if exists trgm_user_book_bookname;

create index trgm_user_book_bookname on user_book using gin (book_name gin_trgm_ops);

drop index if exists trgm_user_book_nickname;

create index trgm_user_book_nickname on user_book using gin (nickname gin_trgm_ops);