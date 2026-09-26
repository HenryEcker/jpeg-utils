package jpegutils

import (
	"bufio"
	"encoding/binary"
	"errors"
	"io"
	"os"
	"strconv"
)

type RGBThumbnailEntry struct {
	R uint8 // 1 Byte Red Value
	G uint8 // 1 Byte Green Value
	B uint8 // 1 Byte Blue Value
}

type RGBThumbnailEntries []RGBThumbnailEntry

// APP0Entry Stores JFIF DATA
type APP0Entry struct {
	Length       uint16              // Length of segment excluding APP0 Marker
	Identifier   [5]byte             // A literal string "JFIF\x00" as 5 bytes
	JFIFVersion  [2]byte             // First Byte is Major Version, Second Byte is Minor Version
	DensityUnits uint8               // 00 for no units, 01 for inches, 02 for centimeters
	XDensity     uint16              // X Axis - Pixels or DPI or DPCM based on Density Units
	YDensity     uint16              // Y Axis - Pixels or DPI or DPCM based on Density Units
	XThumbnail   uint8               // Thumbnail horizontal pixel count
	YThumbnail   uint8               // Thumbnail vertical pixel count
	RGBn         RGBThumbnailEntries // Packed 24-bit RGB Values for thumbnail pixels, length n = XThumbnail * YThumbnail
}

/*****************************    ERRORS    *****************************/

var EofBeforeJFIF = errors.New("reached EOF before parsing JFIF")

type JFIFParseError struct {
	Field string
}

func (e *JFIFParseError) Error() string {
	return "could not parse JFIF field: " + e.Field
}

func NewJFIFParseError(field string) error {
	return &JFIFParseError{field}
}

type JFIFFieldValidationError struct {
	Message string
}

func (e *JFIFFieldValidationError) Error() string {
	return e.Message
}

func NewJFIFFieldValidationError(message string) error {
	return &JFIFFieldValidationError{message}
}

// handleMaybeEOF converts io.EOF into EofBeforeJFIF and passes through all other errors
func handleMaybeEOF(maybeEOF error) error {
	if maybeEOF == io.EOF {
		return EofBeforeJFIF
	}
	return maybeEOF
}

////////////////////////////////////////////////////////////////////////

// Search bufio.Reader for JPEG Markers (FF) and check against second byte (marker)
// to find the correct location to being reading data
func readUntilMarker(bufReader *bufio.Reader, marker byte) (err error) {
	for {
		// All JPEG INDICATORS START WITH FF
		_, err = bufReader.ReadBytes(0xFF)
		if err != nil {
			return handleMaybeEOF(err)
		}
		// Read One More Byte to see which indicator this is
		var bv byte
		bv, err = bufReader.ReadByte()
		if err != nil {
			return handleMaybeEOF(err)
		}
		// Check if the second byte matches the marker provided
		if bv == marker {
			return nil // bufReader is now staged at the correct location
		}
	}
}

// ParseJfifReader reads bufReader and extracts section bytes into struct APP0Entry
func ParseJfifReader(bufReader *bufio.Reader) (*APP0Entry, error) {
	// FF E0 is the JFIF-APP0 marker (All Markers start with FF so only provide second byte)
	err := readUntilMarker(bufReader, 0xE0)
	if err != nil {
		return nil, err
	}

	e := &APP0Entry{}

	err = binary.Read(bufReader, binary.BigEndian, &e.Length)
	if err != nil {
		return nil, NewJFIFParseError("length")
	}

	err = binary.Read(bufReader, binary.BigEndian, &e.Identifier)
	if err != nil {
		return nil, NewJFIFParseError("identifier")
	}
	if string(e.Identifier[:]) != "JFIF\x00" {
		return nil, NewJFIFFieldValidationError("failed to validate JFIF Identifier")
	}

	err = binary.Read(bufReader, binary.BigEndian, &e.JFIFVersion)
	if err != nil {
		return nil, NewJFIFParseError("version")
	}

	err = binary.Read(bufReader, binary.BigEndian, &e.DensityUnits)
	if err != nil {
		return nil, NewJFIFParseError("density units")
	}

	err = binary.Read(bufReader, binary.BigEndian, &e.XDensity)
	if err != nil {
		return nil, NewJFIFParseError("x density")
	}

	err = binary.Read(bufReader, binary.BigEndian, &e.YDensity)
	if err != nil {
		return nil, NewJFIFParseError("y density")
	}

	err = binary.Read(bufReader, binary.BigEndian, &e.XThumbnail)
	if err != nil {
		return nil, NewJFIFParseError("x thumbnail")
	}

	err = binary.Read(bufReader, binary.BigEndian, &e.YThumbnail)
	if err != nil {
		return nil, NewJFIFParseError("y thumbnail")
	}

	// Allocate RGBn Entry
	rgbEntryLength := int(e.XThumbnail * e.YThumbnail)
	e.RGBn = make([]RGBThumbnailEntry, rgbEntryLength)

	// Loop over Thumbnail RGB Entries and build into slice
	for i := range rgbEntryLength {
		rgbEntry := RGBThumbnailEntry{}
		err = binary.Read(bufReader, binary.BigEndian, &rgbEntry.R)
		if err != nil {
			return nil, NewJFIFParseError("RGB Entry " + strconv.Itoa(i) + " R")
		}
		err = binary.Read(bufReader, binary.BigEndian, &rgbEntry.G)
		if err != nil {
			return nil, NewJFIFParseError("RGB Entry " + strconv.Itoa(i) + " G")
		}
		err = binary.Read(bufReader, binary.BigEndian, &rgbEntry.B)
		if err != nil {
			return nil, NewJFIFParseError("RGB Entry " + strconv.Itoa(i) + " B")
		}

		e.RGBn = append(e.RGBn, rgbEntry)
	}

	return e, nil
}

// ParseJfif accepts an imagePath and creates a bufio.Reader for ParseJfifReader
func ParseJfif(imagePath string) (*APP0Entry, error) {
	file, err := os.Open(imagePath)
	if err != nil {
		return nil, err
	}

	defer func(fp *os.File) {
		_ = fp.Close()
	}(file)

	bufReader := bufio.NewReader(file)

	return ParseJfifReader(bufReader)
}
