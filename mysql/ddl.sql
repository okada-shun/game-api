DROP DATABASE IF EXISTS `game_user`;
CREATE DATABASE `game_user`;
USE `game_user`;

DROP TABLE IF EXISTS `game_user`.`users`;
CREATE TABLE IF NOT EXISTS `game_user`.`users`(
  `user_id` CHAR(36) PRIMARY KEY NOT NULL,
  `name` VARCHAR(32) NOT NULL,
  `private_key` VARCHAR(64) NOT NULL
);

DROP TABLE IF EXISTS `game_user`.`rarities`;
CREATE TABLE IF NOT EXISTS `game_user`.`rarities`(
  `id` INT PRIMARY KEY AUTO_INCREMENT NOT NULL,
  `rarity_name` VARCHAR(32) NOT NULL,
  `weight` INT NOT NULL,
  `HPup` INT NOT NULL
);

INSERT INTO rarities(rarity_name, weight, HPup) VALUES ('SR', 1, 1000);
INSERT INTO rarities(rarity_name, weight, HPup) VALUES ('R', 5, 500);
INSERT INTO rarities(rarity_name, weight, HPup) VALUES ('N', 14, 0);

DROP TABLE IF EXISTS `game_user`.`characters`;
CREATE TABLE IF NOT EXISTS `game_user`.`characters`(
  `id` INT PRIMARY KEY AUTO_INCREMENT NOT NULL,
  `character_name` VARCHAR(32) NOT NULL,
  `HP` INT NOT NULL
);

INSERT INTO characters(character_name, HP) VALUES ("Mercury", 1000);
INSERT INTO characters(character_name, HP) VALUES ("Venus", 1000);
INSERT INTO characters(character_name, HP) VALUES ("Earth", 1000);
INSERT INTO characters(character_name, HP) VALUES ("Mars", 1000);
INSERT INTO characters(character_name, HP) VALUES ("Jupiter", 1000);
INSERT INTO characters(character_name, HP) VALUES ("Saturn", 1000);
INSERT INTO characters(character_name, HP) VALUES ("Uranus", 1000);
INSERT INTO characters(character_name, HP) VALUES ("Neptune", 1000);
INSERT INTO characters(character_name, HP) VALUES ("Pluto", 1000);
INSERT INTO characters(character_name, HP) VALUES ("Sun", 1000);

DROP TABLE IF EXISTS `game_user`.`gachas`;
CREATE TABLE IF NOT EXISTS `game_user`.`gachas`(
  `id` INT PRIMARY KEY AUTO_INCREMENT NOT NULL,
  `gacha_name` VARCHAR(32) NOT NULL
);

INSERT INTO gachas(gacha_name) VALUES ("Gacha_A");
INSERT INTO gachas(gacha_name) VALUES ("Gacha_B");
INSERT INTO gachas(gacha_name) VALUES ("Gacha_C");

DROP TABLE IF EXISTS `game_user`.`gacha_characters`;
CREATE TABLE IF NOT EXISTS `game_user`.`gacha_characters`(
  `gacha_character_id` CHAR(36) PRIMARY KEY NOT NULL,
  `gacha_id` INT NOT NULL,
  `character_id` INT NOT NULL,
  `rarity_id` INT NOT NULL,
  `HP` INT NOT NULL
);

INSERT INTO gacha_characters(gacha_character_id, gacha_id, character_id, rarity_id, HP) VALUES (UUID(), 1, 1, 1, 0);
INSERT INTO gacha_characters(gacha_character_id, gacha_id, character_id, rarity_id, HP) VALUES (UUID(), 1, 2, 2, 0);
INSERT INTO gacha_characters(gacha_character_id, gacha_id, character_id, rarity_id, HP) VALUES (UUID(), 1, 3, 2, 0);
INSERT INTO gacha_characters(gacha_character_id, gacha_id, character_id, rarity_id, HP) VALUES (UUID(), 1, 4, 2, 0);
INSERT INTO gacha_characters(gacha_character_id, gacha_id, character_id, rarity_id, HP) VALUES (UUID(), 1, 5, 3, 0);
INSERT INTO gacha_characters(gacha_character_id, gacha_id, character_id, rarity_id, HP) VALUES (UUID(), 1, 6, 3, 0);
INSERT INTO gacha_characters(gacha_character_id, gacha_id, character_id, rarity_id, HP) VALUES (UUID(), 1, 7, 3, 0);
INSERT INTO gacha_characters(gacha_character_id, gacha_id, character_id, rarity_id, HP) VALUES (UUID(), 1, 8, 3, 0);
INSERT INTO gacha_characters(gacha_character_id, gacha_id, character_id, rarity_id, HP) VALUES (UUID(), 1, 9, 3, 0);
INSERT INTO gacha_characters(gacha_character_id, gacha_id, character_id, rarity_id, HP) VALUES (UUID(), 1, 10, 3, 0);

INSERT INTO gacha_characters(gacha_character_id, gacha_id, character_id, rarity_id, HP) VALUES (UUID(), 2, 1, 3, 0);
INSERT INTO gacha_characters(gacha_character_id, gacha_id, character_id, rarity_id, HP) VALUES (UUID(), 2, 2, 3, 0);
INSERT INTO gacha_characters(gacha_character_id, gacha_id, character_id, rarity_id, HP) VALUES (UUID(), 2, 3, 3, 0);
INSERT INTO gacha_characters(gacha_character_id, gacha_id, character_id, rarity_id, HP) VALUES (UUID(), 2, 4, 1, 0);
INSERT INTO gacha_characters(gacha_character_id, gacha_id, character_id, rarity_id, HP) VALUES (UUID(), 2, 5, 2, 0);
INSERT INTO gacha_characters(gacha_character_id, gacha_id, character_id, rarity_id, HP) VALUES (UUID(), 2, 6, 2, 0);
INSERT INTO gacha_characters(gacha_character_id, gacha_id, character_id, rarity_id, HP) VALUES (UUID(), 2, 7, 2, 0);
INSERT INTO gacha_characters(gacha_character_id, gacha_id, character_id, rarity_id, HP) VALUES (UUID(), 2, 8, 3, 0);
INSERT INTO gacha_characters(gacha_character_id, gacha_id, character_id, rarity_id, HP) VALUES (UUID(), 2, 9, 3, 0);
INSERT INTO gacha_characters(gacha_character_id, gacha_id, character_id, rarity_id, HP) VALUES (UUID(), 2, 10, 3, 0);

INSERT INTO gacha_characters(gacha_character_id, gacha_id, character_id, rarity_id, HP) VALUES (UUID(), 3, 1, 3, 0);
INSERT INTO gacha_characters(gacha_character_id, gacha_id, character_id, rarity_id, HP) VALUES (UUID(), 3, 2, 3, 0);
INSERT INTO gacha_characters(gacha_character_id, gacha_id, character_id, rarity_id, HP) VALUES (UUID(), 3, 3, 3, 0);
INSERT INTO gacha_characters(gacha_character_id, gacha_id, character_id, rarity_id, HP) VALUES (UUID(), 3, 4, 3, 0);
INSERT INTO gacha_characters(gacha_character_id, gacha_id, character_id, rarity_id, HP) VALUES (UUID(), 3, 5, 3, 0);
INSERT INTO gacha_characters(gacha_character_id, gacha_id, character_id, rarity_id, HP) VALUES (UUID(), 3, 6, 3, 0);
INSERT INTO gacha_characters(gacha_character_id, gacha_id, character_id, rarity_id, HP) VALUES (UUID(), 3, 7, 1, 0);
INSERT INTO gacha_characters(gacha_character_id, gacha_id, character_id, rarity_id, HP) VALUES (UUID(), 3, 8, 2, 0);
INSERT INTO gacha_characters(gacha_character_id, gacha_id, character_id, rarity_id, HP) VALUES (UUID(), 3, 9, 2, 0);
INSERT INTO gacha_characters(gacha_character_id, gacha_id, character_id, rarity_id, HP) VALUES (UUID(), 3, 10, 2, 0);

CREATE VIEW character_HP AS
SELECT gacha_characters.gacha_character_id, rarities.HPup + characters.HP AS HP
FROM gacha_characters
JOIN characters
ON gacha_characters.character_id = characters.id
JOIN rarities
ON gacha_characters.rarity_id = rarities.id;

UPDATE gacha_characters, character_HP
SET gacha_characters.HP = character_HP.HP
WHERE gacha_characters.gacha_character_id = character_HP.gacha_character_id;

DROP TABLE IF EXISTS `game_user`.`user_characters`;
CREATE TABLE IF NOT EXISTS `game_user`.`user_characters`(
  `user_character_id` CHAR(36) PRIMARY KEY NOT NULL,
  `user_id` VARCHAR(36) NOT NULL,
  `gacha_character_id` VARCHAR(36) NOT NULL
);