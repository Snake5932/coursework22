DROP TYPE if exists genre;

create type genre as enum ('sci-fi', 'fantasy', 'comics', 'satire', 'crime', 'adventure', 'historical',
                           'religious', 'horror', 'nonfiction');