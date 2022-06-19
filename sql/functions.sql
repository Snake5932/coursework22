CREATE OR REPLACE FUNCTION get_admin() returns uuid
LANGUAGE plpgsql
AS $$
declare admin_guid uuid;
BEGIN
    admin_guid = (select guid from administrator
                   where rights & B'10' != B'00'
                   order by check_num
                   limit 1);
    update administrator
    set check_num = check_num + 1
    where guid = admin_guid;
    return admin_guid;
END;
$$;

CREATE OR REPLACE FUNCTION delete_admin() RETURNS trigger
LANGUAGE plpgsql
AS $$
declare checking uuid;
BEGIN
    if old.nickname = 'melkor' then
        RAISE EXCEPTION 'deletion of melkor is prohibited';
    end if;
    return old;
END;
$$;

CREATE OR REPLACE FUNCTION delete_admin2() RETURNS trigger
LANGUAGE plpgsql
AS $$
declare checking uuid;
BEGIN
    if 1 = (select count(*) from administrator) then
        RAISE EXCEPTION 'last admin';
    end if;
    return old;
END;
$$;

CREATE OR REPLACE FUNCTION add_verif() RETURNS trigger
LANGUAGE plpgsql
AS $$
BEGIN
    insert into verification (book_id)
    values(new.guid);
    return new;
END;
$$;

CREATE OR REPLACE FUNCTION delete_user() RETURNS trigger
LANGUAGE plpgsql
AS $$
BEGIN
    RAISE EXCEPTION 'user deletion is prohibited';
END;
$$;

CREATE OR REPLACE FUNCTION delete_verification() RETURNS trigger
LANGUAGE plpgsql
AS $$
BEGIN
    update administrator
    set check_num = check_num - 1
    where guid = old.admin_id;
    return old;
END;
$$;

CREATE OR REPLACE FUNCTION delete_page() RETURNS trigger
LANGUAGE plpgsql
AS $$
BEGIN
    delete from book
    where guid in (select distinct book_id from old_t);
    return NULL;
END;
$$;

CREATE OR REPLACE FUNCTION check_book_page_constr() RETURNS trigger
LANGUAGE plpgsql
AS $$
BEGIN
    if new.guid not in (select distinct book_id from page) then
        RAISE EXCEPTION 'book has no pages';
    end if;
    return new;
END;
$$;

CREATE OR REPLACE FUNCTION check_book_interm_constr() RETURNS trigger
LANGUAGE plpgsql
AS $$
BEGIN
    if new.guid not in (select distinct book_id from aut_book_interm) then
        RAISE EXCEPTION 'book has no author';
    end if;
    return new;
END;
$$;

CREATE OR REPLACE FUNCTION check_author_interm_constr() RETURNS trigger
LANGUAGE plpgsql
AS $$
BEGIN
    if new.guid not in (select distinct author_id from aut_book_interm) then
        RAISE EXCEPTION 'author has no book';
    end if;
    return new;
END;
$$;

CREATE OR REPLACE FUNCTION add_pages() RETURNS trigger
LANGUAGE plpgsql
AS $$
BEGIN
    update book set page_num = page_num + 1 where guid = new.book_id;
    return new;
END;
$$;