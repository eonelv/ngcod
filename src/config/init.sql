CREATE TABLE [t_bd_user] (
  [id] INTEGER NOT NULL PRIMARY KEY,
  [name] TEXT(64) NOT NULL,
  [account] VARCHAR(44) NOT NULL,
  [pass] VARCHAR(44),
  [createtime] TEXT(255));

CREATE UNIQUE INDEX [IDX_USER_ACCOUNT_PASS] ON [t_bd_user] ([account], [pass]);