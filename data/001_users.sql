CREATE TABLE users (
	-- https://stackoverflow.com/questions/7905859/is-there-auto-increment-in-sqlite#answer-7905936
	id INTEGER PRIMARY KEY,
	email TEXT DEFAULT '' NOT NULL,
	name TEXT DEFAULT '' NOT NULL,
	password TEXT DEFAULT '' NOT NULL,
	verified BOOLEAN DEFAULT FALSE NOT NULL
);

CREATE UNIQUE INDEX `idx_email__users` ON `users` (`email`) WHERE `email` != '';
