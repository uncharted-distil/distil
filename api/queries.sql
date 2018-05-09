SELECT
  width_bucket("Games_played", 140, 3562, 50) - 1 as bucket,
  CAST ((width_bucket("Games_played", 140, 3562, 50) - 1) * 68.46 + 140 as double precision) AS "histogram_Games_played",
  COUNT(*) AS count FROM d_185_baseball data INNER JOIN d_185_baseball_result result ON data."d3mIndex" = result.index
WHERE
  result.result_id ='/home/chris/dev/go_workspace/src/github.com/unchartedsoftware/distil/datasets/73896aa8-3f5e-4860-8767-fb0f8dad7ba1-0/tables/learningData.csv'
  -- AND CAST("value" as double precision) IN (*)
GROUP BY
  width_bucket("Games_played", 140, 3562, 50) - 1
ORDER BY
  "histogram_Games_played";



SELECT base."SOUTH", result.value, COUNT(*) AS count
FROM d_534_cps_85_wages_result AS result INNER JOIN d_534_cps_85_wages AS base ON result.index = base."d3mIndex"
WHERE  CAST("value" as double precision) IN ($1) AND result.result_id = $2 and result.target = $3
GROUP BY result.value, base."SOUTH"
ORDER BY count desc;
