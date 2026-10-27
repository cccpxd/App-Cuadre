package services

// ValidationError es un error causado por datos inválidos que envió el
// usuario. La capa HTTP lo responde con 400 y muestra el mensaje tal cual.
// Cualquier otro error (base de datos, red, etc.) es interno: se registra en
// el log y el usuario solo ve un mensaje genérico.
//
// Es un tipo string y no una variable, para no usar variables globales.
type ValidationError string

func (e ValidationError) Error() string { return string(e) }
