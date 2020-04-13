-- SELECT
--   * --   MIN("horsepower") AS "min_horsepower",
--   --   MAX("horsepower") AS "max_horsepower"
-- FROM d_morelearningdata data
-- INNER JOIN d_morelearningdata_result result ON data."d3mIndex" = result.index
-- WHERE
--   result.result_id = '/home/chris/outputs/supporting_files/9dd5f02c-7945-43a4-a89e-b126cf0cf63b_outputs.0.csv'
--   AND "horsepower" != 'NaN';
SELECT
  *
FROM d_morelearningdata_result data
LIMIT
  1000
