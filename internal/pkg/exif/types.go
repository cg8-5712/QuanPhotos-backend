package exif

import "time"

// Data represents parsed EXIF information
type Data struct {
	// Camera info
	CameraMake   string
	CameraModel  string
	SerialNumber string

	// Lens info
	LensMake        string
	LensModel       string
	FocalLength     string
	FocalLength35mm string

	// Shooting parameters
	Aperture        string
	ShutterSpeed    string
	ISO             int
	ExposureMode    string
	ExposureProgram string
	MeteringMode    string
	WhiteBalance    string
	Flash           string
	ExposureBias    string

	// Time and location
	TakenAt      *time.Time
	GPSLatitude  *float64
	GPSLongitude *float64
	GPSAltitude  *float64

	// Image info
	ImageWidth  int
	ImageHeight int
	Orientation int
	ColorSpace  string
	Software    string
}

// HasGPS returns true if GPS coordinates are available
func (d *Data) HasGPS() bool {
	return d.GPSLatitude != nil && d.GPSLongitude != nil
}

// HasTakenAt returns true if taken time is available
func (d *Data) HasTakenAt() bool {
	return d.TakenAt != nil
}

// ExposurePrograms maps EXIF exposure program values to human-readable strings
var ExposurePrograms = map[int]string{
	0: "Not defined",
	1: "Manual",
	2: "Normal program",
	3: "Aperture priority",
	4: "Shutter priority",
	5: "Creative program",
	6: "Action program",
	7: "Portrait mode",
	8: "Landscape mode",
}

// MeteringModes maps EXIF metering mode values to human-readable strings
var MeteringModes = map[int]string{
	0:   "Unknown",
	1:   "Average",
	2:   "Center-weighted average",
	3:   "Spot",
	4:   "Multi-spot",
	5:   "Pattern",
	6:   "Partial",
	255: "Other",
}

// WhiteBalances maps EXIF white balance values to human-readable strings
var WhiteBalances = map[int]string{
	0: "Auto",
	1: "Manual",
}

// FlashModes maps EXIF flash values to human-readable strings
var FlashModes = map[int]string{
	0x0:  "No Flash",
	0x1:  "Fired",
	0x5:  "Fired, Return not detected",
	0x7:  "Fired, Return detected",
	0x8:  "On, Did not fire",
	0x9:  "On, Fired",
	0xd:  "On, Return not detected",
	0xf:  "On, Return detected",
	0x10: "Off, Did not fire",
	0x14: "Off, Did not fire, Return not detected",
	0x18: "Auto, Did not fire",
	0x19: "Auto, Fired",
	0x1d: "Auto, Fired, Return not detected",
	0x1f: "Auto, Fired, Return detected",
	0x20: "No flash function",
	0x30: "Off, No flash function",
	0x41: "Fired, Red-eye reduction",
	0x45: "Fired, Red-eye reduction, Return not detected",
	0x47: "Fired, Red-eye reduction, Return detected",
	0x49: "On, Red-eye reduction",
	0x4d: "On, Red-eye reduction, Return not detected",
	0x4f: "On, Red-eye reduction, Return detected",
	0x50: "Off, Red-eye reduction",
	0x58: "Auto, Did not fire, Red-eye reduction",
	0x59: "Auto, Fired, Red-eye reduction",
	0x5d: "Auto, Fired, Red-eye reduction, Return not detected",
	0x5f: "Auto, Fired, Red-eye reduction, Return detected",
}

// ExposureModes maps EXIF exposure mode values to human-readable strings
var ExposureModes = map[int]string{
	0: "Auto",
	1: "Manual",
	2: "Auto bracket",
}

// ColorSpaces maps EXIF color space values to human-readable strings
var ColorSpaces = map[int]string{
	1:     "sRGB",
	2:     "Adobe RGB",
	65535: "Uncalibrated",
}
