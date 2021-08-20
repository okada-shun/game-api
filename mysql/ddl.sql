DROP DATABASE IF EXISTS `game_user`;
CREATE DATABASE `game_user`;
USE `game_user`;

CREATE USER 'okada'@'localhost' IDENTIFIED BY 'password';
GRANT ALL PRIVILEGES ON game_user.* TO 'okada'@'localhost';
SHOW GRANTS FOR 'okada'@'localhost';

DROP TABLE IF EXISTS `game_user`.`users`;
CREATE TABLE IF NOT EXISTS `game_user`.`users`(
  `id` INT PRIMARY KEY AUTO_INCREMENT,
  `name` VARCHAR(32) NOT NULL
);

DROP TABLE IF EXISTS `game_user`.`characters`;
CREATE TABLE IF NOT EXISTS `game_user`.`characters`(
  `id` INT PRIMARY KEY AUTO_INCREMENT,
  `name` VARCHAR(32) NOT NULL,
  `rarity` VARCHAR(32) NOT NULL
);

INSERT INTO characters(name, rarity) VALUES ('Alice', 3);
INSERT INTO characters(name, rarity) VALUES ('Bob', 2);
INSERT INTO characters(name, rarity) VALUES ('Carol', 1);

DROP TABLE IF EXISTS `game_user`.`rarities`;
CREATE TABLE IF NOT EXISTS `game_user`.`rarities`(
  `id` INT PRIMARY KEY AUTO_INCREMENT,
  `name` VARCHAR(32) NOT NULL,
  `weight` INT NOT NULL
);

INSERT INTO rarities(name, weight) VALUES ('N', 60);
INSERT INTO rarities(name, weight) VALUES ('R', 30);
INSERT INTO rarities(name, weight) VALUES ('SR', 10);

DROP TABLE IF EXISTS `game_user`.`user_characters`;
CREATE TABLE IF NOT EXISTS `game_user`.`user_characters`(
  `id` INT PRIMARY KEY AUTO_INCREMENT,
  `user_id` VARCHAR(32) NOT NULL,
  `character_id` VARCHAR(32) NOT NULL
);