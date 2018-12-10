package block

import(
	"crypto/sha256"
	"bytes"
)
//implements the Content interface provided by merkletree and represents the content stored in the tree.
type BlockContent struct {
	X []byte
}
  
//CalculateHash hashes the values of a TestContent
func (t BlockContent) CalculateHash() ([]byte, error) {
	h := sha256.New()
	if _, err := h.Write(t.X); err != nil {
		return nil, err
	}

	return h.Sum(nil), nil
}
  
//Equals tests for equality of two Contents
func (t BlockContent) Equals(other Content) (bool, error) {
	return bytes.Compare(t.X, other.(BlockContent).X) == 0, nil
}



//TestContent implements the Content interface provided by merkletree and represents the content stored in the tree.
type TestContent struct {
	X string
  }
  
  //CalculateHash hashes the values of a TestContent
  func (t TestContent) CalculateHash() ([]byte, error) {
	h := sha256.New()
	if _, err := h.Write([]byte(t.X)); err != nil {
	  return nil, err
	}
  
	return h.Sum(nil), nil
  }
  //Equals tests for equality of two Contents
	func (t TestContent) Equals(other Content) (bool, error) {
		return t.X == other.(TestContent).X, nil
}