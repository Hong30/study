CREATE TABLE `weibos` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `user_id` int(11) NOT NULL,
  `account` varchar(16) COLLATE utf8_bin NOT NULL,
  `content` varchar(64) COLLATE utf8_bin NOT NULL,
  `like_num` int(11) NOT NULL,
  `created_at` int(11) NOT NULL DEFAULT '0',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COLLATE=utf8_bin;