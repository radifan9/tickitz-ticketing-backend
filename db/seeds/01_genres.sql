-- INSERT INTO public.genres (id,"name") 
-- OVERRIDING SYSTEM VALUE
-- VALUES
-- 	 (1,'Action'),
-- 	 (2,'Drama'),
-- 	 (3,'Crime'),
-- 	 (4,'Fantasy'),
-- 	 (5,'Adventure'),
-- 	 (6,'Thriller'),
-- 	 (7,'Sci-Fi'),
-- 	 (59,'Gore');

INSERT INTO public.genres (id, name) 
OVERRIDING SYSTEM VALUE
VALUES
    (28, 'Action'),
    (12, 'Adventure'),
    (16, 'Animation'),
    (35, 'Comedy'),
    (80, 'Crime'),
    (99, 'Documentary'),
    (18, 'Drama'),
    (10751, 'Family'),
    (14, 'Fantasy'),
    (36, 'History'),
    (27, 'Horror'),
    (10402, 'Music'),
    (9648, 'Mystery'),
    (10749, 'Romance'),
    (878, 'Science Fiction'),
    (10770, 'TV Movie'),
    (53, 'Thriller'),
    (10752, 'War'),
    (37, 'Western');
SELECT setval('genres_id_seq', (SELECT MAX(id) FROM genres));