create table  users (
  id INTEGER PRIMARY KEY,
  login varchar(255),
  fullName varchar (255),
  password varchar(255),
  avatar varchar(255)
);

create  table  votes (
  id INTEGER PRIMARY KEY,
  timestamp  long,
  voter varchar(255),
  subject varchar(255)
);

dupa8
dupa8