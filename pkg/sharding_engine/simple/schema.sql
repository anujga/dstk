drop table partition;

CREATE TABLE partition
(
    id          bigint,
    modified_on timestamp without time zone,
    worker_id   bigint,
    start       bytea,
    "end"       bytea,
    url         text,
    desired_state text,
    current_state text default '',
    leader_id    bigint default 0,
    proxy_to bigint[] DEFAULT array[]::bigint[]
);

-- Id: 1
-- Url: localhost:9099
-- Parts:
--   - Start: ""
-- End: "o"
--   - Start: "w"
-- End: "z"

-- drop table partition

insert into partition (id, modified_on, worker_id, start, "end", url)
values (1, now(), 1, E'\\x00', E'\\x1000', 'localhost:6011');

insert into partition (id, modified_on, worker_id, start, "end", url, desired_state, current_state, leader_id)
values (102, now(), 1, E'\\x0000', E'\\x1000', 'localhost:6011', 'follower', 'follower', 1);

insert into partition (id, modified_on, worker_id, start, "end", url, desired_state, leader_id)
values (1, now(), 1, E'\\x0000', E'\\x6000', 'localhost:6011', 'primary', 2);


update partition set proxy_to='{101,102}', desired_state='proxy' where id=1;

update partition set url='localhost:6011' where url='localhost:9099';

update partition set proxy_to='{101, 102}' where id=1;

update partition set current_state='primary' where 1=1;

select * from partition;

delete from partition where modified_on='2020-07-26 05:36:34.104066';
