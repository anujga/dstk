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
    leader_id    bigint,
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
