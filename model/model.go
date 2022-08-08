package model

type Phone string //27123456789 (11 digits or 0+9 -> 27+9)

type Permission string //enum...

type TableName string

type Date string //CCYY-MM-DD SAST

type Coordinates string //<lat>;<lon> both as decimal values

type StringList string //<string>|<string>|... used for tags etc

type Unit string //enum e.g. Kg, Dozen, L, units (just counted items), ...
