DROP DATABASE IF EXISTS `game_user`;
CREATE DATABASE `game_user`;
USE `game_user`;

DROP TABLE IF EXISTS `game_user`.`users`;
CREATE TABLE IF NOT EXISTS `game_user`.`users`(
  `user_id` CHAR(36) PRIMARY KEY NOT NULL,
  `name` VARCHAR(32) NOT NULL
);

DROP TABLE IF EXISTS `game_user`.`rarities`;
CREATE TABLE IF NOT EXISTS `game_user`.`rarities`(
  `rarity_id` INT PRIMARY KEY AUTO_INCREMENT NOT NULL,
  `rarity_name` VARCHAR(32) NOT NULL,
  `weight` INT NOT NULL
);

INSERT INTO rarities(rarity_name, weight) VALUES ('SR', 1);
INSERT INTO rarities(rarity_name, weight) VALUES ('R', 5);
INSERT INTO rarities(rarity_name, weight) VALUES ('N', 14);

DROP TABLE IF EXISTS `game_user`.`gacha_characters`;
CREATE TABLE IF NOT EXISTS `game_user`.`gacha_characters`(
  `character_id` CHAR(36) PRIMARY KEY NOT NULL,
  `gacha_id` INT NOT NULL,
  `rarity_id` INT NOT NULL,
  `character_name` VARCHAR(32) NOT NULL,
  `HP` INT NOT NULL
);

INSERT INTO gacha_characters(character_id, gacha_id, rarity_id, character_name, HP) VALUES (UUID(), 1, 1, "Mercury", 1200);
INSERT INTO gacha_characters(character_id, gacha_id, rarity_id, character_name, HP) VALUES (UUID(), 1, 2, "Venus", 850);
INSERT INTO gacha_characters(character_id, gacha_id, rarity_id, character_name, HP) VALUES (UUID(), 1, 2, "Earth", 800);
INSERT INTO gacha_characters(character_id, gacha_id, rarity_id, character_name, HP) VALUES (UUID(), 1, 2, "Mars", 750);
INSERT INTO gacha_characters(character_id, gacha_id, rarity_id, character_name, HP) VALUES (UUID(), 1, 3, "Jupiter", 450);
INSERT INTO gacha_characters(character_id, gacha_id, rarity_id, character_name, HP) VALUES (UUID(), 1, 3, "Saturn", 425);
INSERT INTO gacha_characters(character_id, gacha_id, rarity_id, character_name, HP) VALUES (UUID(), 1, 3, "Uranus", 400);
INSERT INTO gacha_characters(character_id, gacha_id, rarity_id, character_name, HP) VALUES (UUID(), 1, 3, "Neptune", 400);
INSERT INTO gacha_characters(character_id, gacha_id, rarity_id, character_name, HP) VALUES (UUID(), 1, 3, "Pluto", 375);
INSERT INTO gacha_characters(character_id, gacha_id, rarity_id, character_name, HP) VALUES (UUID(), 1, 3, "Sun", 350);

INSERT INTO gacha_characters(character_id, gacha_id, rarity_id, character_name, HP) VALUES (UUID(), 2, 3, "Mercury", 450);
INSERT INTO gacha_characters(character_id, gacha_id, rarity_id, character_name, HP) VALUES (UUID(), 2, 3, "Venus", 400);
INSERT INTO gacha_characters(character_id, gacha_id, rarity_id, character_name, HP) VALUES (UUID(), 2, 3, "Earth", 350);
INSERT INTO gacha_characters(character_id, gacha_id, rarity_id, character_name, HP) VALUES (UUID(), 2, 1, "Mars", 1250);
INSERT INTO gacha_characters(character_id, gacha_id, rarity_id, character_name, HP) VALUES (UUID(), 2, 2, "Jupiter", 850);
INSERT INTO gacha_characters(character_id, gacha_id, rarity_id, character_name, HP) VALUES (UUID(), 2, 2, "Saturn", 800);
INSERT INTO gacha_characters(character_id, gacha_id, rarity_id, character_name, HP) VALUES (UUID(), 2, 2, "Uranus", 750);
INSERT INTO gacha_characters(character_id, gacha_id, rarity_id, character_name, HP) VALUES (UUID(), 2, 3, "Neptune", 425);
INSERT INTO gacha_characters(character_id, gacha_id, rarity_id, character_name, HP) VALUES (UUID(), 2, 3, "Pluto", 400);
INSERT INTO gacha_characters(character_id, gacha_id, rarity_id, character_name, HP) VALUES (UUID(), 2, 3, "Sun", 375);

INSERT INTO gacha_characters(character_id, gacha_id, rarity_id, character_name, HP) VALUES (UUID(), 3, 3, "Mercury", 450);
INSERT INTO gacha_characters(character_id, gacha_id, rarity_id, character_name, HP) VALUES (UUID(), 3, 3, "Venus", 425);
INSERT INTO gacha_characters(character_id, gacha_id, rarity_id, character_name, HP) VALUES (UUID(), 3, 3, "Earth", 400);
INSERT INTO gacha_characters(character_id, gacha_id, rarity_id, character_name, HP) VALUES (UUID(), 3, 3, "Mars", 400);
INSERT INTO gacha_characters(character_id, gacha_id, rarity_id, character_name, HP) VALUES (UUID(), 3, 3, "Jupiter", 375);
INSERT INTO gacha_characters(character_id, gacha_id, rarity_id, character_name, HP) VALUES (UUID(), 3, 3, "Saturn", 350);
INSERT INTO gacha_characters(character_id, gacha_id, rarity_id, character_name, HP) VALUES (UUID(), 3, 1, "Uranus", 1150);
INSERT INTO gacha_characters(character_id, gacha_id, rarity_id, character_name, HP) VALUES (UUID(), 3, 2, "Neptune", 850);
INSERT INTO gacha_characters(character_id, gacha_id, rarity_id, character_name, HP) VALUES (UUID(), 3, 2, "Pluto", 800);
INSERT INTO gacha_characters(character_id, gacha_id, rarity_id, character_name, HP) VALUES (UUID(), 3, 2, "Sun", 750);

DROP TABLE IF EXISTS `game_user`.`user_characters`;
CREATE TABLE IF NOT EXISTS `game_user`.`user_characters`(
  `user_character_id` CHAR(36) PRIMARY KEY NOT NULL,
  `user_id` VARCHAR(36) NOT NULL,
  `character_id` VARCHAR(36) NOT NULL
);