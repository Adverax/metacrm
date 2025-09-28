package purifiers

type PNGPurifier struct {
	storage ChunkStorage
	next    Purifier
}

func NewPNGPurifier(storage ChunkStorage, next Purifier) *PNGPurifier {
	return &PNGPurifier{
		storage: storage,
		next:    next,
	}
}

func (that *PNGPurifier) Purify(original, derivative string) string {
	if isPNG(derivative) {
		return that.storage.Save(derivative)
	}

	if that.next == nil {
		return derivative
	}

	return that.next.Purify(original, derivative)
}
