drop materialized view if exists admin_book;
drop materialized view if exists user_book;

create materialized view admin_book as
select b.book_name, b.guid, v.admin_id
    from book as b
    inner join verification as v
    on b.guid = v.book_id;

create materialized view user_book as
select u.nickname, b.user_id, b.guid, b.book_name, b.book_genre, b.write_year, b.page_num, a.author_name
    from end_user as u
    inner join book as b
    on b.user_id = u.guid
    inner join aut_book_interm as i
    on b.guid = i.book_id
    inner join author as a
    on a.guid = i.author_id;