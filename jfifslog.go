package jpegutils

import (
	"fmt"
	"log/slog"
)

func (e RGBThumbnailEntry) LogValue() slog.Value {
	return slog.GroupValue(
		slog.Uint64("R", uint64(e.R)),
		slog.Uint64("G", uint64(e.G)),
		slog.Uint64("B", uint64(e.B)),
	)
}

func (rte RGBThumbnailEntries) LogValue() slog.Value {
	values := make([]slog.Value, len(rte))
	for _, u := range rte {
		values = append(values, u.LogValue())
	}
	return slog.AnyValue(values)
}

func (a APP0Entry) LogValue() slog.Value {
	return slog.GroupValue(
		slog.String("Length", fmt.Sprintf("%d", a.Length)),
		slog.String("Identifier", string(a.Identifier[:])),
		slog.String("JFIFVersion", fmt.Sprintf(
			"%02x (%d.%d)",
			a.JFIFVersion,    // Display the Binary in Hex
			a.JFIFVersion[0], // Get First Byte for Major
			a.JFIFVersion[1], // Get Second Byte for Minor
		)),
		slog.String("DensityUnits", fmt.Sprintf("%02x", a.DensityUnits)),
		slog.String("XDensity", fmt.Sprintf("%d", a.XDensity)),
		slog.String("YDensity", fmt.Sprintf("%d", a.YDensity)),
		slog.String("XThumbnail", fmt.Sprintf("%d", a.XThumbnail)),
		slog.String("YThumbnail", fmt.Sprintf("%d", a.YThumbnail)),
		slog.Any("RGBThumbnailEntry", a.RGBn),
	)
}
