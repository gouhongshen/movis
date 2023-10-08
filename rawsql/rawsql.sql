

-- 每分钟事务平均耗时

SELECT
    DATE_FORMAT(response_at, '%Y-%m-%d %H:%i') AS min_timestamp,
    AVG(duration) / CAST(1000000000 AS UNSIGNED) AS avg_duration
FROM statement_info
WHERE transaction_id IN (
    SELECT DISTINCT JSON_UNQUOTE(JSON_EXTRACT(extra, '$.txn_id'))
    FROM span_info
    WHERE span_name = "MysqlCmdExecutor.executeStmt"
)
  AND statement LIKE "%__mo_stmt_id%"
GROUP BY min_timestamp
Order By min_timestamp asc;



-- 每分钟 S3FS.Read 耗时
SELECT
    DATE_FORMAT(end_time, '%Y-%m-%d %H:%i') AS min_timestamp,
    AVG(duration) / CAST(1000000000 AS UNSIGNED) AS avg_duration
FROM span_info
WHERE trace_id IN (
    SELECT DISTINCT trace_id
    FROM span_info
    WHERE span_name = "MysqlCmdExecutor.executeStmt"
)
  AND span_name = "LocalFS.Read"
GROUP BY min_timestamp
Order By min_timestamp asc;



-- 每分钟 S3FS.read 流量 （MB 每秒）
# 流量 = 字节数 / 时间
SELECT
    DATE_FORMAT(end_time, '%Y-%m-%d %H:%i') AS min_timestamp,
    SUM(CAST(json_unquote(json_extract(extra, "$.size")) as UNSIGNED)) / CAST(1024*1024*60 as UNSIGNED) AS kb_by_sec
FROM span_info
WHERE span_name = "LocalFS.read"
GROUP BY min_timestamp
Order By min_timestamp asc;



-- 每分钟 Compile.Run 耗时
SELECT
    DATE_FORMAT(end_time, '%Y-%m-%d %H:%i') AS min_timestamp,
    AVG(duration) / CAST(1000000000 AS UNSIGNED) AS avg_duration
FROM span_info
WHERE trace_id IN (
    SELECT DISTINCT trace_id
    FROM span_info
    WHERE span_name = "MysqlCmdExecutor.executeStmt"
    ) AND span_name = "Compile.Run"
GROUP BY min_timestamp
Order By min_timestamp asc;



# 每分钟 rollback 次数

SELECT
    DATE_FORMAT(end_time, '%Y-%m-%d %H:%i') AS min_timestamp,
    count(*) as rollback_cnt
FROM span_info
WHERE span_name = "Session.RollbackTxn" AND node_uuid="37386263-6331-6466-6661-666364623064"
GROUP BY min_timestamp
Order By min_timestamp asc;



-- 事务的总耗时

-- select SUM(duration) / CAST(1000000000 AS UNSIGNED) as avg_dur
-- from statement_info
-- where transaction_id IN (
--     SELECT DISTINCT JSON_UNQUOTE(JSON_EXTRACT(extra, '$.txn_id'))
--     FROM span_info
--     WHERE span_name = "Compile.Run"
-- )
--     AND statement LIKE "%__mo_stmt_id%" ;




-- 每分钟平均缓存命中率
SELECT
    DATE_FORMAT(`timestamp`, '%Y-%m-%d %H:%i') AS min_timestamp,
    AVG(CAST(JSON_UNQUOTE(JSON_EXTRACT(extra, '$."FileService Cache Hit Rate"')) AS DOUBLE)) AS avg_cache_hit_rate
FROM log_info
GROUP BY min_timestamp
ORDER BY min_timestamp ASC;




-- 每分钟内存平均缓存命中率
-- SELECT
--     DATE_FORMAT(`timestamp`, '%Y-%m-%d %H:%i') AS min_timestamp,
--     AVG(CAST(JSON_UNQUOTE(JSON_EXTRACT(extra, '$."FileService Cache Memory Hit Rate"')) AS DOUBLE)) AS avg_cache_hit_rate
-- FROM log_info
-- WHERE `timestamp` < "2023-09-25 07:16:00" and `timestamp` >  "2023-09-25 06:45:56"
-- GROUP BY min_timestamp
-- ORDER BY min_timestamp ASC;




-- 每分钟磁盘平均缓存命中率
SELECT
    DATE_FORMAT(`timestamp`, '%Y-%m-%d %H:%i') AS min_timestamp,
    AVG(CAST(JSON_UNQUOTE(JSON_EXTRACT(extra, '$."FileService Cache Disk Hit Rate"')) AS DOUBLE)) AS avg_cache_hit_rate
FROM log_info
GROUP BY min_timestamp
ORDER BY min_timestamp ASC;




-- S3FS.Read 占 Compile.Run 的 比例

-- select SUM(duration) / CAST(1000000000 AS UNSIGNED) as total_dur
-- from span_info where span_name = "S3FS.Read" and trace_id in (
-- 	select trace_id from span_info where span_name = "Compile.Run" and JSON_UNQUOTE(JSON_EXTRACT(extra, '$.statement')) LIKE "%__mo_stmt_id%"
-- )

-- 9901.912819888

-- select SUM(duration) / CAST(1000000000 AS UNSIGNED) as total_dur
-- from  span_info where span_name = "Compile.Run" and JSON_UNQUOTE(JSON_EXTRACT(extra, '$.statement')) LIKE "%__mo_stmt_id%"

-- 42891.906269315




-- executeStmt 的总用时
-- select SUM(duration) / CAST(1000000000 AS UNSIGNED) as total_dur
-- from span_info where span_name = "MysqlCmdExecutor.executeStmt" and trace_id in (
-- 	select trace_id from span_info where span_name = "Compile.Run" and JSON_UNQUOTE(JSON_EXTRACT(extra, '$.statement')) LIKE "%__mo_stmt_id%"
-- )

-- 54299.824955776


-- MysqlCmdExecutor.doComQuery 的总用时
-- select SUM(duration) / CAST(1000000000 AS UNSIGNED) as total_dur
-- from span_info where span_name = "MysqlCmdExecutor.doComQuery" and trace_id in (
-- 	select trace_id from span_info where span_name = "Compile.Run" and JSON_UNQUOTE(JSON_EXTRACT(extra, '$.statement')) LIKE "%__mo_stmt_id%"
-- )







-- 按name聚合 object 访问次数

select
    count(*) as frequency, json_unquote(json_extract(extra, '$.name')) as obj_name, min(end_time), max(end_time)
from span_info
where span_name="LocalFS.read" and json_unquote(json_extract(extra, '$.name')) not like "%.csv%"
group by obj_name
order by frequency desc;


-- select count(distinct json_unquote(json_extract(extra, '$.name')))
-- from span_info
-- where span_name="S3FS.read" and node_type="CN" and span_kind="s3FSOperation" and json_unquote(json_extract(extra, '$.name')) not like "%.csv%"





-- 每分钟 executeStmt 平均耗时

SELECT
    DATE_FORMAT(start_time, '%Y-%m-%d %H:%i') AS min_timestamp,
    AVG(duration) / CAST(1000000000 AS UNSIGNED) AS avg_duration
FROM span_info
WHERE span_name = "MysqlCmdExecutor.executeStmt"
GROUP BY min_timestamp
Order By min_timestamp asc;





-- select count( trace_id)
-- from span_info
-- WHERE span_name = "MysqlCmdExecutor.doComQuery" and trace_id in (
-- 	select trace_id from span_info where span_name="MysqlCmdExecutor.executeStmt"
-- )

-- select count( trace_id)
-- from span_info where span_name = "MysqlCmdExecutor.doComQuery"



-- S3FS.Read 前十分钟的总延迟
-- explain SELECT
-- 	SUM(duration) / CAST(1000000000 AS UNSIGNED) as total_dur
-- FROM system.span_info
-- WHERE span_name = "S3FS.Read" and trace_id  IN (
--     SELECT DISTINCT trace_id
--     FROM system.span_info
--     WHERE span_name = "Compile.runOnce"
-- )

-- 前十分钟：37000.507978302
-- 总共：2628.657351012


-- S3FS.Read 前十分钟的总数据量
-- SELECT
-- 	cast(JSON_UNQUOTE(JSON_EXTRACT(extra, '$.size')) as unsigned)
-- FROM system.span_info
-- WHERE span_name = "S3FS.read" and trace_id IN (
--     SELECT DISTINCT trace_id
--     FROM system.span_info
--     WHERE span_name = "Compile.runOnce"
-- )

--  S3FS.Read: 1002646442099
--  S3FS.read: 14500724993



-- S3FS.Read 总数据量
--  SELECT
-- 	SUM(cast(JSON_UNQUOTE(JSON_EXTRACT(extra, '$.size')) as unsigned))
-- FROM span_info
-- WHERE trace_id IN (
--     SELECT DISTINCT trace_id
--     FROM span_info
--     WHERE span_name = "Compile.runOnce"
-- )
--     AND span_name = "S3FS.read"

--  S3FS.Read: 19487667176951
--  S3FS.read: 2733232998


-- Error Code: 0 Server sent unknown charsetnr (0) . Please report




# 每个 TPCC session 中各部分时间占比
# 各个 Session 总时间
select SUM(duration)/CAST(1000*1000*1000 AS UNSIGNED) as total_sec, count(*) as cnt
from system.span_info
where span_name = "Routine.handleRequest" and trace_id in (
    select trace_id from system.span_info
    where span_name = "Routine.handleRequest"
    group by trace_id
    order by count(trace_id) desc limit 100
)
group by trace_id
order by trace_id desc;


# 各个 阶段 总时间
select SUM(duration)/CAST(1000*1000*1000 AS UNSIGNED) as total_sec, count(*) as cnt
from system.span_info
where span_name = "TxnHandler.CommitTxn" and trace_id in (
    select trace_id from system.span_info
    where span_name = "Routine.handleRequest"
    group by trace_id
    order by count(trace_id) desc limit 100
)
group by trace_id
order by trace_id desc;


/*

RoutineManager.Handler --> Routine.handleRequest
									|
									|
									\
				MysqlCmdExecutor.ExecRequest --> MysqlCmdExecutor.doComQuery
														|
                                                        |
                                                        \
												   executeStmt --> TxnComputationWrapper.Compile
														|
														|
														\ --> Compile.Run --> Compile.ScopeRun --> S3FS.Read

*/



# metadata scan, 一张表有多少个 object，多少个 blk，大小分别是多少

# 以object分组
select
    count(*) as cnt,
    sum(origin_size) / cast(1024 * 1024 as BIGINT) as sum_origin_size,
    sum(compress_size) / cast(1024 * 1024 as BIGINT) as sum_compress_size
from metadata_scan('system.rawlog', '*') g
group by object_name;

select count(*), avg(sum_origin_size), avg(sum_compress_size)
from (
    select sum(origin_size) / cast(1024 * 1024 as BIGINT) as sum_origin_size,
           sum(compress_size) / cast(1024 * 1024 as BIGINT) as sum_compress_size
    from metadata_scan('system.rawlog', '*') g
    group by object_name
     );


select
    count(*) as cnt,
    sum(origin_size) / cast(1024 * 1024 as BIGINT) as sum_origin_size,
    sum(compress_size) / cast(1024 * 1024 as BIGINT) as sum_compress_size
from metadata_scan('system.rawlog', '*') t
group by block_id;



# 一个 object 有多少行
select avg(cnt), count(*)
from (select sum(rows_cnt) as cnt
      from metadata_scan('system.statement_info', '*') t
      group by object_name);

select count(*) from mo_catalog.mo_tables;