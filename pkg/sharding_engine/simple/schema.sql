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
values (1, now(), 1, '\\x', '\\x9', 'localhost:9099');

insert into partition (id, modified_on, worker_id, start, "end", url)
values (2, now(), 1, '\\x9', '\\xFFFFFFFFF', 'localhost:9099');