DROP DATABASE IF EXISTS `game_user`;
CREATE DATABASE `game_user`;
USE `game_user`;

DROP TABLE IF EXISTS `game_user`.`users`;
CREATE TABLE IF NOT EXISTS `game_user`.`users`(
  `id` INT PRIMARY KEY AUTO_INCREMENT,
  `user_id` VARCHAR(100) NOT NULL,
  `name` VARCHAR(32) NOT NULL
);

DROP TABLE IF EXISTS `game_user`.`characters`;
CREATE TABLE IF NOT EXISTS `game_user`.`characters`(
  `id` INT PRIMARY KEY AUTO_INCREMENT,
  `character_id` VARCHAR(100) NOT NULL,
  `name` VARCHAR(32) NOT NULL,
  `weight` INT NOT NULL
);

INSERT INTO characters(character_id, name, weight) VALUES (UUID(), 'Alice_SR', 10);
INSERT INTO characters(character_id, name, weight) VALUES (UUID(), 'Bob_R', 30);
INSERT INTO characters(character_id, name, weight) VALUES (UUID(), 'Bob_R', 30);
INSERT INTO characters(character_id, name, weight) VALUES (UUID(), 'Bob_R', 30);
INSERT INTO characters(character_id, name, weight) VALUES (UUID(), 'Carol_N', 60);
INSERT INTO characters(character_id, name, weight) VALUES (UUID(), 'Carol_N', 60);
INSERT INTO characters(character_id, name, weight) VALUES (UUID(), 'Carol_N', 60);
INSERT INTO characters(character_id, name, weight) VALUES (UUID(), 'Carol_N', 60);
INSERT INTO characters(character_id, name, weight) VALUES (UUID(), 'Carol_N', 60);
INSERT INTO characters(character_id, name, weight) VALUES (UUID(), 'Carol_N', 60);

DROP TABLE IF EXISTS `game_user`.`user_characters`;
CREATE TABLE IF NOT EXISTS `game_user`.`user_characters`(
  `id` INT PRIMARY KEY AUTO_INCREMENT,
  `user_character_id` VARCHAR(100) NOT NULL,
  `user_id` VARCHAR(100) NOT NULL,
  `character_id` VARCHAR(100) NOT NULL
);