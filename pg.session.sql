-- SELECT MIN("coordinates" [1]) AS "min_x",
--     MAX("coordinates" [5]) AS "max_x",
--     MIN("coordinates" [2]) AS "min_y",
--     MAX("coordinates" [6]) AS "max_y"
-- FROM d_bigearth_tiny;
SELECT "coordinates[1]"
FROM d_bigearth_tiny;
