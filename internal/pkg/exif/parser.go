package exif

import (
	"fmt"
	"io"
	"math"
	"time"

	"github.com/rwcarlsen/goexif/exif"
	"github.com/rwcarlsen/goexif/tiff"
)

// Parser parses EXIF data from images
type Parser struct{}

// NewParser creates a new EXIF parser
func NewParser() *Parser {
	return &Parser{}
}

// Parse extracts EXIF data from an image reader
func (p *Parser) Parse(r io.Reader) (*Data, error) {
	x, err := exif.Decode(r)
	if err != nil {
		// Return empty data if no EXIF found (not an error)
		if err == io.EOF {
			return &Data{}, nil
		}
		return &Data{}, nil
	}

	data := &Data{}

	// Camera info
	data.CameraMake = p.getString(x, exif.Make)
	data.CameraModel = p.getString(x, exif.Model)
	// Note: BodySerialNumber is not available in standard EXIF tags
	// data.SerialNumber = p.getString(x, exif.BodySerialNumber)

	// Lens info
	data.LensMake = p.getString(x, exif.LensMake)
	data.LensModel = p.getString(x, exif.LensModel)
	data.FocalLength = p.getFocalLength(x)
	data.FocalLength35mm = p.getFocalLength35mm(x)

	// Shooting parameters
	data.Aperture = p.getAperture(x)
	data.ShutterSpeed = p.getShutterSpeed(x)
	data.ISO = p.getISO(x)
	data.ExposureMode = p.getExposureMode(x)
	data.ExposureProgram = p.getExposureProgram(x)
	data.MeteringMode = p.getMeteringMode(x)
	data.WhiteBalance = p.getWhiteBalance(x)
	data.Flash = p.getFlash(x)
	data.ExposureBias = p.getExposureBias(x)

	// Time
	data.TakenAt = p.getDateTime(x)

	// GPS
	lat, lon, err := x.LatLong()
	if err == nil {
		data.GPSLatitude = &lat
		data.GPSLongitude = &lon
	}
	data.GPSAltitude = p.getGPSAltitude(x)

	// Image info
	data.ImageWidth = p.getInt(x, exif.ImageWidth)
	data.ImageHeight = p.getInt(x, exif.ImageLength)
	data.Orientation = p.getInt(x, exif.Orientation)
	data.ColorSpace = p.getColorSpace(x)
	data.Software = p.getString(x, exif.Software)

	return data, nil
}

// getString extracts a string value from EXIF tag
func (p *Parser) getString(x *exif.Exif, tag exif.FieldName) string {
	t, err := x.Get(tag)
	if err != nil {
		return ""
	}
	val, err := t.StringVal()
	if err != nil {
		return ""
	}
	return val
}

// getInt extracts an integer value from EXIF tag
func (p *Parser) getInt(x *exif.Exif, tag exif.FieldName) int {
	t, err := x.Get(tag)
	if err != nil {
		return 0
	}
	val, err := t.Int(0)
	if err != nil {
		return 0
	}
	return val
}

// getRational extracts a rational value from EXIF tag
func (p *Parser) getRational(x *exif.Exif, tag exif.FieldName) (num, denom int64, ok bool) {
	t, err := x.Get(tag)
	if err != nil {
		return 0, 0, false
	}
	n, d, err := t.Rat2(0)
	if err != nil {
		return 0, 0, false
	}
	return n, d, true
}

// getAperture extracts and formats aperture value
func (p *Parser) getAperture(x *exif.Exif) string {
	num, denom, ok := p.getRational(x, exif.FNumber)
	if !ok || denom == 0 {
		return ""
	}
	fNumber := float64(num) / float64(denom)
	return fmt.Sprintf("f/%.1f", fNumber)
}

// getShutterSpeed extracts and formats shutter speed value
func (p *Parser) getShutterSpeed(x *exif.Exif) string {
	num, denom, ok := p.getRational(x, exif.ExposureTime)
	if !ok || denom == 0 {
		return ""
	}

	if num == 0 {
		return ""
	}

	// If exposure time < 1 second, show as fraction
	if num < denom {
		// Simplify fraction
		gcd := gcd(num, denom)
		num /= gcd
		denom /= gcd
		return fmt.Sprintf("%d/%d s", num, denom)
	}

	// Otherwise show as decimal
	seconds := float64(num) / float64(denom)
	if seconds >= 1 {
		return fmt.Sprintf("%.1f s", seconds)
	}
	return fmt.Sprintf("1/%d s", int(1/seconds))
}

// getFocalLength extracts and formats focal length
func (p *Parser) getFocalLength(x *exif.Exif) string {
	num, denom, ok := p.getRational(x, exif.FocalLength)
	if !ok || denom == 0 {
		return ""
	}
	fl := float64(num) / float64(denom)
	return fmt.Sprintf("%.0f mm", fl)
}

// getFocalLength35mm extracts and formats 35mm equivalent focal length
func (p *Parser) getFocalLength35mm(x *exif.Exif) string {
	val := p.getInt(x, exif.FocalLengthIn35mmFilm)
	if val == 0 {
		return ""
	}
	return fmt.Sprintf("%d mm", val)
}

// getISO extracts ISO value
func (p *Parser) getISO(x *exif.Exif) int {
	t, err := x.Get(exif.ISOSpeedRatings)
	if err != nil {
		return 0
	}
	val, err := t.Int(0)
	if err != nil {
		return 0
	}
	return val
}

// getExposureMode extracts and formats exposure mode
func (p *Parser) getExposureMode(x *exif.Exif) string {
	val := p.getInt(x, exif.ExposureMode)
	if mode, ok := ExposureModes[val]; ok {
		return mode
	}
	return ""
}

// getExposureProgram extracts and formats exposure program
func (p *Parser) getExposureProgram(x *exif.Exif) string {
	val := p.getInt(x, exif.ExposureProgram)
	if program, ok := ExposurePrograms[val]; ok {
		return program
	}
	return ""
}

// getMeteringMode extracts and formats metering mode
func (p *Parser) getMeteringMode(x *exif.Exif) string {
	val := p.getInt(x, exif.MeteringMode)
	if mode, ok := MeteringModes[val]; ok {
		return mode
	}
	return ""
}

// getWhiteBalance extracts and formats white balance
func (p *Parser) getWhiteBalance(x *exif.Exif) string {
	val := p.getInt(x, exif.WhiteBalance)
	if wb, ok := WhiteBalances[val]; ok {
		return wb
	}
	return ""
}

// getFlash extracts and formats flash status
func (p *Parser) getFlash(x *exif.Exif) string {
	val := p.getInt(x, exif.Flash)
	if flash, ok := FlashModes[val]; ok {
		return flash
	}
	return ""
}

// getExposureBias extracts and formats exposure bias
func (p *Parser) getExposureBias(x *exif.Exif) string {
	t, err := x.Get(exif.ExposureBiasValue)
	if err != nil {
		return ""
	}

	if t.Format() == tiff.RatVal {
		num, denom, err := t.Rat2(0)
		if err != nil || denom == 0 {
			return ""
		}
		bias := float64(num) / float64(denom)
		if bias == 0 {
			return "0 EV"
		}
		if bias > 0 {
			return fmt.Sprintf("+%.1f EV", bias)
		}
		return fmt.Sprintf("%.1f EV", bias)
	}

	return ""
}

// getDateTime extracts photo taken time
func (p *Parser) getDateTime(x *exif.Exif) *time.Time {
	t, err := x.DateTime()
	if err != nil {
		return nil
	}
	return &t
}

// getGPSAltitude extracts GPS altitude
func (p *Parser) getGPSAltitude(x *exif.Exif) *float64 {
	t, err := x.Get(exif.GPSAltitude)
	if err != nil {
		return nil
	}

	num, denom, err := t.Rat2(0)
	if err != nil || denom == 0 {
		return nil
	}

	alt := float64(num) / float64(denom)

	// Check altitude reference (0 = above sea level, 1 = below)
	ref, err := x.Get(exif.GPSAltitudeRef)
	if err == nil {
		refVal, err := ref.Int(0)
		if err == nil && refVal == 1 {
			alt = -alt
		}
	}

	return &alt
}

// getColorSpace extracts and formats color space
func (p *Parser) getColorSpace(x *exif.Exif) string {
	val := p.getInt(x, exif.ColorSpace)
	if cs, ok := ColorSpaces[val]; ok {
		return cs
	}
	return ""
}

// gcd calculates greatest common divisor
func gcd(a, b int64) int64 {
	a = int64(math.Abs(float64(a)))
	b = int64(math.Abs(float64(b)))
	for b != 0 {
		a, b = b, a%b
	}
	return a
}
