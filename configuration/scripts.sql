DROP TABLE IF EXISTS `items`;
CREATE TABLE `items` (
  `item_id` int(11) NOT NULL AUTO_INCREMENT,
  `instrument` varchar(50) NOT NULL,
  `item_key` varchar(50) NOT NULL,
  `direction` varchar(10) NOT NULL,
  `qty` int NOT NULL,
  `price` decimal(12,4) DEFAULT NULL,
  PRIMARY KEY (`item_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;