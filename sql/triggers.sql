drop trigger if exists del_admin on administrator;

create trigger del_admin after delete
    on administrator
    for each row
    execute function delete_admin();
    
drop trigger if exists add_verification on book;

create trigger add_verification after insert
    on book
    for each row
    execute function add_verif();
    
drop trigger if exists user_del on end_user;

create trigger user_del before delete
    on end_user
    for each statement
    execute function delete_user();
    
drop trigger if exists verif_del on verification;

create trigger verif_del after delete
    on verification
    for each row
    execute function delete_verification();
    
drop trigger if exists page_del on page;
    
create trigger page_del after delete
    on page
    referencing old table as old_t
    for each statement
    execute function delete_page();
    
drop trigger if exists book_page on book;

create constraint trigger book_page after insert
    on book
    deferrable initially deferred
    for each row
    execute function check_book_page_constr();
    
drop trigger if exists book_interm on book;

create constraint trigger book_interm after insert
    on book
    deferrable initially deferred
    for each row
    execute function check_book_interm_constr();
    
drop trigger if exists author_interm on author;

create constraint trigger author_interm after insert
    on author
    deferrable initially deferred
    for each row
    execute function check_author_interm_constr();
    
drop trigger if exists add_page_num on page;

create constraint trigger add_page_num after insert
    on page
    deferrable initially deferred
    for each row
    execute function add_pages();