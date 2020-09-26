create table requests
(
    id      bigserial primary key,
    method  varchar(80),
    url     varchar(80),
    headers varchar(80),
    body    varchar(80)
);

-- pg_ctl -D /usr/local/var/postgres start
