# version 1.0
# author Julian.Yue
CREATE DATABASE IF NOT EXISTS mikudos_message_pusher;
USE mikudos_message_pusher;

# channel message
# DROP TABLE channel_msg;
CREATE TABLE IF NOT EXISTS channel_msg (
	id bigint unsigned NOT NULL AUTO_INCREMENT PRIMARY KEY, # auto increment id
	skey varchar(64) NOT NULL, # channelId key
	mid bigint unsigned NOT NULL, # message id
	ttl bigint NOT NULL, # message expire second
	msg blob NOT NULL, # message content
	ctime TIMESTAMP DEFAULT CURRENT_TIMESTAMP, # message create time
	mtime TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP, # message update time
	INDEX ix_channel_msg_1 (skey),
	INDEX ix_channel_msg_2 (skey, mid),
	INDEX ix_channel_msg_3 (ttl)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
