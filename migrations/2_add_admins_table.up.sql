create table if not exists admins (
    id integer primary key ,
    user_id integer not null ,
    foreign key (user_id) references users(id)
);