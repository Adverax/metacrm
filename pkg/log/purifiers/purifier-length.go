package purifiers

type LenPurifier struct {
	maxLen  int
	storage ChunkStorage
	next    Purifier
}

func NewLenPurifier(storage ChunkStorage, maxLength int, next Purifier) *LenPurifier {
	return &LenPurifier{
		maxLen:  maxLength,
		storage: storage,
		next:    next,
	}
}

func (that *LenPurifier) Purify(original, derivative string) string {
	if len(derivative) <= that.maxLen {
		if that.next == nil {
			return derivative
		}

		return that.next.Purify(original, derivative)
	}

	return that.storage.Save(derivative)
}
