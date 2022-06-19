CREATE OR REPLACE FUNCTION check_upd_uuid() returns trigger
LANGUAGE plpgsql
AS $$
BEGIN
    RAISE EXCEPTION 'uuid update is prohibited';
END;
$$;

CREATE OR REPLACE FUNCTION prohibit_user_gen() RETURNS trigger
LANGUAGE plpgsql
AS $$
BEGIN
    RAISE EXCEPTION 'direct interactions with user_gen are prohibited';
END;
$$;

create trigger check_uuid_book before update
    of guid, user_id on book
    for each row
    execute function check_upd_uuid();
    
create trigger check_uuid_page before update
    of guid, book_id on page
    for each row
    execute function check_upd_uuid();
    
create trigger check_uuid_author before update
    of guid on author
    for each row
    execute function check_upd_uuid();
    
create trigger check_uuid_aut_book_interm before update
    of author_id, book_id on aut_book_interm
    for each row
    execute function check_upd_uuid();
    
create trigger check_uuid_verification before update
    of guid, book_id on verification
    for each row
    execute function check_upd_uuid();
    
create trigger check_uuid_administrator before update
    of guid on administrator
    for each row
    execute function check_upd_uuid();
    
create trigger check_uuid_end_user before update
    of guid on end_user
    for each row
    execute function check_upd_uuid();

create trigger user_gen_insert before insert
    on user_gen
    for each statement
    execute function prohibit_user_gen();
    
create trigger user_gen_update before update
    on user_gen
    for each statement
    execute function prohibit_user_gen();
    
create trigger user_gen_delete before delete
    on user_gen
    for each statement
    execute function prohibit_user_gen();