CREATE TABLE departure_profiles(
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  destination_id BIGINT NOT NULL,
  departure_time TIME NOT NULL,
  preparation_time INTEGER NOT NULL,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

  CONSTRAINT fk_departure_profiles_destinations 
      FOREIGN KEY (destination_id) REFERENCES destinations(id)
      ON DELETE CASCADE 
      ON UPDATE CASCADE
);