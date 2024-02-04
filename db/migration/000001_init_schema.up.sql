CREATE TABLE IF NOT EXISTS tables (
    id INT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    size INT NOT NULL,
    occupied INT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
) ENGINE=INNODB;

CREATE TABLE IF NOT EXISTS guests (
    id INT NOT NULL auto_increment PRIMARY KEY,
    guest_name VARCHAR(255) NOT NULL,
    entourage INT NOT NULL,
    table_id INT NOT NULL,
    arrival_time TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    FOREIGN KEY (table_id)
        REFERENCES tables (id)
        ON UPDATE RESTRICT ON DELETE CASCADE
) ENGINE=INNODB;


CREATE TABLE IF NOT EXISTS arrivals (
    id INT NOT NULL auto_increment PRIMARY KEY,
    guest_id INT NOT NULL,
    table_id INT NOT NULL,
    party_size INT NOT NULL,

    FOREIGN KEY (table_id)
        REFERENCES tables (id)
        ON UPDATE RESTRICT ON DELETE CASCADE,

    FOREIGN KEY (guest_id)
        REFERENCES guests (id)
        ON UPDATE RESTRICT ON DELETE CASCADE
) ENGINE =INNODB;