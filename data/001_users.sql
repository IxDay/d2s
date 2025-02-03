CREATE TABLE users (
	-- https://stackoverflow.com/questions/7905859/is-there-auto-increment-in-sqlite#answer-7905936
	id INTEGER PRIMARY KEY,
	email TEXT DEFAULT '' NOT NULL,
	name TEXT DEFAULT '' NOT NULL,
	created DATETIME,
	UNIQUE(email)
);

CREATE UNIQUE INDEX `idx_email__users` ON `users` (`email`) WHERE `email` != '';
