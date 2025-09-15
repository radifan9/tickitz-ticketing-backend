INSERT INTO public.movies (id,title,synopsis,poster_img,backdrop_img,duration_minutes,release_date,director_id,age_rating_id,created_at,updated_at,archived_at) 
OVERRIDING SYSTEM VALUE
VALUES
	 (4,'Avatar','A Marine on Pandora becomes torn between following orders and protecting his new home.','avatar.jpg','avatar_bg.jpg',162,'2009-12-18',4,1,'2025-09-07 13:20:29.219947+07','2025-09-07 13:20:29.219947+07',NULL),
	 (5,'The Lord of the Rings: The Return of the King','The final confrontation against Sauron begins.','lotr.jpg','lotr_bg.jpg',201,'2003-12-17',5,2,'2025-09-07 13:20:29.219947+07','2025-09-07 13:20:29.219947+07',NULL),
	 (6,'Dune: Part Two','Paul Atreides unites with the Fremen to wage war against House Harkonnen.','dune2.jpg','dune2_bg.jpg',165,'2025-11-14',6,NULL,'2025-09-07 13:27:51.554373+07','2025-09-07 13:27:51.554373+07',NULL),
	 (7,'The Batman: Part II','Batman faces new challenges in Gotham after the rise of a powerful enemy.','batman2.jpg','batman2_bg.jpg',160,'2026-10-03',7,NULL,'2025-09-07 13:27:51.554373+07','2025-09-07 13:27:51.554373+07',NULL),
	 (8,'Wonder Woman 3','Diana Prince confronts a mystical force threatening humanity.','ww3.jpg','ww3_bg.jpg',145,'2027-06-18',8,NULL,'2025-09-07 13:27:51.554373+07','2025-09-07 13:27:51.554373+07',NULL),
	 (9,'Thor: Legacy','Thor faces the consequences of his past while protecting the cosmos.','thor_legacy.jpg','thor_legacy_bg.jpg',150,'2026-05-05',9,NULL,'2025-09-07 13:27:51.554373+07','2025-09-07 13:27:51.554373+07',NULL),
	 (10,'Gladiator II','Lucius, the son of Lucilla, must fight for Rome''s future.','gladiator2.jpg','gladiator2_bg.jpg',155,'2025-11-22',10,NULL,'2025-09-07 13:27:51.554373+07','2025-09-07 13:27:51.554373+07',NULL),
	 (2,'The Shawshank Redemption','Two men form a bond while serving time in prison.','shawshank.jpg','shawshank_bg.jpg',142,'1994-09-23',2,2,'2025-09-07 13:20:29.219947+07','2025-09-08 08:07:13.497849+07',NULL),
	 (3,'Avengers: Endgame','The Avengers assemble once more to undo Thanos'' snap.','endgame.jpg','endgame_bg.jpg',181,'2019-04-26',3,2,'2025-09-07 13:20:29.219947+07','2025-09-08 11:45:24.788199+07',NULL),
	 (1,'Inception','A thief who enters the dreams of others to steal secrets.','inception.jpg','inception_bg.jpg',148,'2010-07-16',1,2,'2025-09-07 13:20:29.219947+07','2025-09-09 19:09:52.570633+07','2025-09-09 19:09:52.570633+07'),
	 (12,'Wonder Woman 4','','ww3.jpg','',145,'2028-06-18',40,3,'2025-09-14 21:31:06.583978+07','2025-09-14 21:31:06.583978+07',NULL);
