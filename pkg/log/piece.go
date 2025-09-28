package log

import "context"

type Piece struct {
	entry *Entry
}

func (that *Piece) WithField(key string, value interface{}) LoggerPiece {
	peace := that.entry.Logger.peaces.Get()
	defer that.entry.Logger.peaces.Put(peace)

	peace.entry = that.entry.withField(key, value)
	return peace
}

func (that *Piece) WithFields(fields Fields) LoggerPiece {
	peace := that.entry.Logger.peaces.Get()
	defer that.entry.Logger.peaces.Put(peace)

	peace.entry = that.entry.withFields(fields)
	return peace
}

func (that *Piece) WithError(err error) LoggerPiece {
	peace := that.entry.Logger.peaces.Get()
	defer that.entry.Logger.peaces.Put(peace)

	peace.entry = that.entry.withError(err)
	return peace
}

func (that *Piece) Message(ctx context.Context, msg string) {
	that.entry.Log(ctx, that.entry.Level, msg)
}

func (that *Piece) Messagef(ctx context.Context, format string, args ...interface{}) {
	that.entry.Logf(ctx, that.entry.Level, format, args...)
}
