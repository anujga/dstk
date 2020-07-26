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

delete from partition where 1=1;

insert into partition (id, modified_on, worker_id, start, "end", url, desired_state)
values (1, now(), 1, E'\\x0000', E'\\x6000', 'localhost:6011', 'primary');

insert into partition (id, modified_on, worker_id, start, "end", url, desired_state, leader_id)
values (11, now(), 1, E'\\x0000', E'\\x2000', 'localhost:6011', 'catchingup', 1);
insert into partition (id, modified_on, worker_id, start, "end", url, desired_state, leader_id)
values (12, now(), 1, E'\\x2000', E'\\x6000', 'localhost:6011', 'catchingup', 1);
-- what if follower is received first
update partition set desired_state='follower' where id=11 or id=12;
update partition set proxy_to='{11,12}', desired_state='proxy' where id=1;
update partition set desired_state='primary' where id=11 or id=12;
update partition set desired_state='proxy' where id=1;
update partition set desired_state='retired' where id=1;


select * from partition;



update partition set desired_state='proxy' where id=1;

update partition set url='localhost:6011' where url='localhost:9099';

update partition set proxy_to='{101, 102}' where id=1;


update partition set current_state='' where 1=1;

