CREATE TABLE partition
(
    id          bigint,
    modified_on timestamp without time zone,
    worker_id   bigint,
    start       bytea,
    "end"       bytea,
    url         text
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
values (1, now(), 1, E'\\x00', E'\\x90', 'localhost:9099');

insert into partition (id, modified_on, worker_id, start, "end", url)
values (2, now(), 1, E'\\x90', E'\\xFFFFFFFFFF', 'localhost:9099');
