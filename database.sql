create table studios (
    id serial primary key,
    name varchar(256) not null,
    abbreviated varchar(50) default ''::character varying,
    type integer default 1
);
INSERT INTO studios (name, abbreviated, type)
VALUES ('Lostfilm', 'lf', 2);


create table videos(
    id serial primary key,
    name text default ''
);

create table files (
    id serial primary key,
    filepath varchar(100) not null,
    video_id integer references videos(id)
);
