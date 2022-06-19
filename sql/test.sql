insert into administrator (nickname, email, pass_hash, rights)
values('melkor', 'shtjnk@yandex.ru', 'xxxx', B'11'), ('sauron', 'shtjnk91@gmail.com', 'xxxx', B'11'),
('morgoth', 'shtjnk111@gmail.com', 'xxxx', B'11');

insert into end_user (nickname, email, pass_hash)
values('melkor', 'shtjnk@yandex.ru', 'xxxx'), ('sauron', 'shtjnk91@gmail.com', 'xxxx');

select * from end_user;

insert into book (user_id, book_name, book_genre)
values('780818a9-a74a-4d07-ae2d-753fc9f0e8be', 'lord of the rings', 'fantasy'),
('780818a9-a74a-4d07-ae2d-753fc9f0e8be', 'hobbit', 'fantasy'),
('780818a9-a74a-4d07-ae2d-753fc9f0e8be', 'dune', 'sci-fi'),
('780818a9-a74a-4d07-ae2d-753fc9f0e8be', 'star-wars', 'sci-fi');
                   
select distinct guid, user_id, book_name, book_genre, write_year from book;

delete from administrator
where nickname = 'morgoth';

select * from verification;

delete from book where book_name = 'hobbit';

select * from book;

select * from administrator;

select * from verification;

delete from verification where guid='2dc97363-6e3a-4df6-95b7-493752fbe86b';

insert into verification(book_id, admin_id) values('202c0b18-2038-46c2-b90c-fe655c10bf14', '090621d2-f23c-480b-be25-bab319960cf0');
