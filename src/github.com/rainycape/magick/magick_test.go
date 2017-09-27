package magick

import (
	"bytes"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

type Fataler interface {
	Fatalf(string, ...interface{})
}

func decodeFile(fat Fataler, name string) *Image {
	f, err := os.Open(filepath.Join("test_data", name))
	if err != nil {
		fat.Fatalf("error reading %s: %s", name, err)
	}
	defer f.Close()
	im, err := Decode(f)
	if err != nil {
		fat.Fatalf("error decoding %s: %s", name, err)
	}
	if im == nil {
		fat.Fatalf("%s: no image", name)
	}
	return im
}

func encodeFile(t *testing.T, name string, im *Image, info *Info) {
	f, err := os.OpenFile(filepath.Join("test_data", "out."+name), os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	err = im.Encode(f, info)
	if err != nil {
		t.Fatal(err)
	}
}

func testImage(t *testing.T, im *Image, frames, width, height, depth int, format string) {
	if im.NFrames() != frames {
		t.Errorf("Invalid number of frames, expected %d, got %d", frames, im.NFrames())
	}
	if im.Width() != width || im.Height() != height {
		t.Errorf("Invalid image dimensions, expected %dx%d, got %dx%d", width, height, im.Width(), im.Height())
	}
	if im.Format() != format {
		t.Errorf("Invalid image format, expected %s got %s", format, im.Format())
	}
	if im.Depth() != depth {
		t.Errorf("Invalid depth format, expected %v got %v", depth, im.Depth())
	}
	if im.NFrames() > 1 {
		_, err := im.Coalesce()
		if err != nil {
			t.Errorf("error coalescing: %s", err)
		}
	}
}

func recodeImage(t *testing.T, im *Image, info *Info) *Image {
	buf := &bytes.Buffer{}
	err := im.Encode(buf, info)
	if err != nil {
		t.Fatal(err)
	}
	decoded, err := DecodeData(buf.Bytes())
	if err != nil {
		t.Fatal(err)
	}
	return decoded
}

func testEntropy(t *testing.T, name string, entropy float32) {
	im := decodeFile(t, name)
	if e := im.Entropy(); e != entropy {
		// IM and GM implementations might return slightly different results
		// due to differences in the precision used during calculations
		delta := e - entropy
		if delta > -0.0001 && delta < 0.0001 {
			t.Logf("Slightly different entropy (due to different precision) for %s. Expected %v, got %v (delta %v)", name, entropy, e, delta)
		} else {
			t.Errorf("Invalid entropy for %s: expecting %v, got %v", name, entropy, e)
		}
	}
}

func TestInfo(t *testing.T) {
	t.Logf("Using backend %s", Backend())
	if formats, err := SupportedFormats(); err != nil {
		t.Error(err)
	} else {
		t.Logf("%d supported formats %v", len(formats), formats)
	}
}

func TestDecode(t *testing.T) {
	im := decodeFile(t, "wizard.png")
	testImage(t, im, 1, 1104, 1468, 8, "PNG")
	cloned, err := im.Clone()
	if err != nil {
		t.Fatal(err)
	}
	testImage(t, cloned, 1, 1104, 1468, 8, "PNG")
	anim := decodeFile(t, "Newtons_cradle_animation_book_2.gif")
	testImage(t, anim, 36, 480, 360, 8, "GIF")
	im2 := decodeFile(t, "lenna.jpg")
	testImage(t, im2, 1, 512, 512, 8, "JPEG")
}

func TestResize(t *testing.T) {
	im1 := decodeFile(t, "wizard.png")
	res1, err := im1.Resize(500, 600, FQuadratic)
	if err != nil {
		t.Fatal(err)
	}
	testImage(t, res1, 1, 500, 600, 8, "PNG")
	im2 := decodeFile(t, "Newtons_cradle_animation_book_2.gif")
	testImage(t, im2, 36, 480, 360, 8, "GIF")
	res2, err := im2.Resize(240, 180, FQuadratic)
	if err != nil {
		t.Fatal(err)
	}
	testImage(t, res2, 36, 240, 180, 8, "GIF")
}

func TestGif(t *testing.T) {
	im := decodeFile(t, "Newtons_cradle_animation_book_2.gif")
	testImage(t, im, 36, 480, 360, 8, "GIF")
	encodeFile(t, "newton.gif", im, nil)
	cloned, err := im.Clone()
	if err != nil {
		t.Fatal(err)
	}
	testImage(t, cloned, 36, 480, 360, 8, "GIF")
	encodeFile(t, "newton2.gif", cloned, nil)
	frame1, err := im.Frame(0)
	if err != nil {
		t.Fatal(err)
	}
	testImage(t, frame1, 1, 480, 360, 8, "GIF")
	cframe1, err := frame1.Clone()
	if err != nil {
		t.Fatal(err)
	}
	testImage(t, cframe1, 1, 480, 360, 8, "GIF")
}

func TestGifEncode(t *testing.T) {
	im := decodeFile(t, "Newtons_cradle_animation_book_2.gif")
	testImage(t, im, 36, 480, 360, 8, "GIF")
	coalesced, err := im.Coalesce()
	if err != nil {
		t.Fatal(err)
	}
	data, err := coalesced.GifEncode()
	if err != nil {
		t.Fatal(err)
	}
	im2, err := DecodeData(data)
	if err != nil {
		t.Fatal(err)
	}
	testImage(t, im2, 36, 480, 360, 8, "GIF")
}

func TestEncode(t *testing.T) {
	im1 := decodeFile(t, "wizard.png")
	res1, err := im1.Resize(500, 600, FQuadratic)
	if err != nil {
		t.Fatal(err)
	}
	im2 := recodeImage(t, res1, nil)
	testImage(t, im2, 1, 500, 600, 8, "PNG")
	info := NewInfo()
	info.SetFormat("JPEG")
	im3 := recodeImage(t, res1, info)
	testImage(t, im3, 1, 500, 600, 8, "JPEG")

	gif1 := decodeFile(t, "Newtons_cradle_animation_book_2.gif")
	gif2 := recodeImage(t, gif1, nil)
	if gif1.Duration() != gif2.Duration() {
		t.Errorf("Invalid duration, expected %d, got %d", gif1.Duration(), gif2.Duration())
	}
	testImage(t, gif2, 36, 480, 360, 8, "GIF")
	gif3, err := gif2.Resize(240, 180, FQuadratic)
	if err != nil {
		t.Fatal(err)
	}
	gif4 := recodeImage(t, gif3, nil)
	testImage(t, gif4, 36, 240, 180, 8, "GIF")
	encodeFile(t, "newton-small.gif", gif4, nil)
	if gif1.Duration() != gif4.Duration() {
		t.Errorf("Invalid duration, expected %d, got %d", gif1.Duration(), gif4.Duration())
	}
	info.SetFormat("PNG")
	nongif := recodeImage(t, gif3, info)
	testImage(t, nongif, 1, 240, 180, 8, "PNG")
	encodeFile(t, "newton.png", nongif, nil)
}

func TestList(t *testing.T) {
	list1 := decodeFile(t, "Newtons_cradle_animation_book_2.gif")
	nframes := list1.NFrames()
	list1.Append(list1)
	if list1.NFrames() != 2*nframes {
		t.Errorf("Error appending self, expected %d frames, got %d %p", 2*nframes, list1.NFrames(), list1)
	}
	nframes = list1.NFrames()
	img1 := decodeFile(t, "wizard.png")
	list1.Append(img1)
	if list1.NFrames() != nframes+1 {
		t.Errorf("Error appending single, expected %d frames, got %d", nframes+1, list1.NFrames())
	}
	nframes = list1.NFrames()
	list1.RemoveFirst()
	if list1.NFrames() != nframes-1 {
		t.Errorf("Error removing first: expected %d frames, got %d", nframes-1, list1.NFrames())
	}
	nframes = list1.NFrames()
	list1.RemoveLast()
	if list1.NFrames() != nframes-1 {
		t.Errorf("Error removing last: expected %d frames, got %d", nframes-1, list1.NFrames())
	}
	nframes = list1.NFrames()
	frame, err := list1.Frame(5)
	if err != nil {
		t.Fatal(err)
	}
	frame.Remove()
	if list1.NFrames() != nframes-1 {
		t.Errorf("Error removing specific frame: expected %d frames, got %d", nframes-1, list1.NFrames())
	}
}

func TestEntropy(t *testing.T) {
	testEntropy(t, "wizard.png", 5.073119)
	testEntropy(t, "lenna.jpg", 8.774539)
}

func TestDecodeOptimized(t *testing.T) {
	im := decodeFile(t, "optimized.gif")
	testImage(t, im, 10, 651, 721, 8, "GIF")
	// This one requires gifsicle with --unoptimize
	im2 := decodeFile(t, "math.gif")
	testImage(t, im2, 158, 500, 350, 8, "GIF")
	// This one requires gifsicle with --unoptimize and piping via convert if using GM
	im3 := decodeFile(t, "kick_grandma.gif")
	testImage(t, im3, 25, 240, 180, 8, "GIF")
	im4 := decodeFile(t, "seCq.gif")
	testImage(t, im4, 136, 270, 165, 8, "GIF")
}

func TestDecodeVideoGif(t *testing.T) {
	im := decodeFile(t, "football.gif")
	testImage(t, im, 347, 312, 232, 8, "GIF")
}

func TestCropResizeAnimated(t *testing.T) {
	im := decodeFile(t, "optimized.gif")
	res, err := im.CropResize(100, 100, FHamming, CSCenter)
	if err != nil {
		t.Fatal(err)
	}
	testImage(t, res, 10, 100, 100, 8, "GIF")
}

func TestAverage(t *testing.T) {
	im := decodeFile(t, "lenna.jpg")
	avg, err := im.AverageColor()
	if err != nil {
		t.Fatal(err)
	}
	if avg.Red != 133 || avg.Green != 80 || avg.Blue != 68 {
		t.Errorf("expected (133, 80, 68), got %+v instead", *avg)
	}
}

func TestProperties(t *testing.T) {
	im := decodeFile(t, "wizard.png")
	count := len(im.Properties())
	if count == 0 {
		// not testing exact number since it varies depending
		// on backend.
		t.Fatal("expecting some properties, got zero")
	}
	if !im.SetProperty("go", "go") {
		t.Error("property not set")
	}
	if len(im.Properties()) != count+1 {
		t.Errorf("expecting %d properties, got %d instead", count+1, len(im.Properties()))
	}
	goVal := im.Property("go")
	if goVal != "go" {
		t.Errorf("expecting property go = \"go\", got go = %q", goVal)
	}
	if !im.SetProperty("go", "2go") {
		t.Error("property not set")
	}
	if !im.HasProperty("go") {
		t.Error("should have property")
	}
	goVal = im.Property("go")
	if goVal != "2go" {
		t.Errorf("expecting property go = \"2go\", got go = %q", goVal)
	}
	if !im.RemoveProperty("go") {
		t.Error("property not removed")
	}
	if im.HasProperty("go") {
		t.Error("should not have property")
	}
	goVal = im.Property("go")
	if len(goVal) != 0 {
		t.Errorf("expecting property go = \"\", got go = %q", goVal)
	}
	if im.RemoveProperty("go") {
		t.Error("property removed when not present")
	}
	im.DestroyProperties()
	count = len(im.Properties())
	if count != 0 {
		t.Errorf("expecting no properties, got %d instead", count)
	}
}

func TestQuality(t *testing.T) {
	im := decodeFile(t, "wizard.png")
	var buf1 bytes.Buffer
	var buf2 bytes.Buffer
	info := NewInfo()
	info.SetFormat("JPEG")
	info.SetQuality(10)
	if err := im.Encode(&buf1, info); err != nil {
		t.Fatal(err)
	}
	info.SetQuality(100)
	if err := im.Encode(&buf2, info); err != nil {
		t.Fatal(err)
	}
	if buf2.Len() <= buf1.Len() {
		t.Errorf("quality = 100 generates %d bytes, quality = 10, %d - first should be bigger", buf2.Len(), buf1.Len())
	}
}

func TestNewImage(t *testing.T) {
	red := decodeFile(t, "red.png")
	im, err := New(red.Width(), red.Height())
	if err != nil {
		t.Fatal(err)
	}
	im.Composite(CompositeCopy, red, 0, 0)
	stats, err := red.Compare(im)
	if err != nil {
		t.Fatal(err)
	}
	if !stats.IsZero() {
		t.Errorf("painted image not equal to source: %v", stats)
	}
}

func TestCompare(t *testing.T) {
	white := decodeFile(t, "white.png")
	red := decodeFile(t, "red.png")
	eq, err := white.IsEqual(white)
	if err != nil {
		t.Fatal(err)
	}
	if !eq {
		t.Errorf("image should be equal to itself")
	}
	eq, err = white.IsEqual(red)
	if err != nil {
		t.Fatal(err)
	}
	if eq {
		t.Errorf("image can't be equal to red")
	}
}

func TestChannels(t *testing.T) {
	white := decodeFile(t, "white.png")
	red := decodeFile(t, "red.png")
	wRed, err := red.ChannelImage(CRed)
	if err != nil {
		t.Fatal(err)
	}
	stats, err := white.Compare(wRed)
	if err != nil {
		t.Fatal(err)
	}
	if !stats.IsZero() {
		t.Errorf("repeated red channel is not equal to white image: %+v", stats)
	}
}

func TestPHash(t *testing.T) {
	red := decodeFile(t, "red.png")
	phash1, err := red.PHash()
	if err != nil {
		t.Fatal(err)
	}
	redSmall := decodeFile(t, "red-small.png")
	phash2, err := redSmall.PHash()
	if err != nil {
		t.Fatal(err)
	}
	if phash1 != phash2 {
		t.Errorf("red.png and red-small.png have different phash: %v and %v (delta %v)", phash1, phash2, phash1.Compare(phash2))
	}
	t.Logf("red.png PHASH = %v (%v)", phash1, uint64(phash1))
	lenna := decodeFile(t, "lenna.jpg")
	phash3, err := lenna.PHash()
	if err != nil {
		t.Fatal(err)
	}
	lennaSmall := decodeFile(t, "lenna-small.jpg")
	phash4, err := lennaSmall.PHash()
	if err != nil {
		t.Fatal(err)
	}
	if phash3 != phash4 {
		t.Errorf("lenna.jpg and lenna-small.jpg have different phash: %v and %v (delta %v)", phash3, phash4, phash3.Compare(phash4))
	}
	t.Logf("lenna.jpg PHASH = %v (%v)", phash3, uint64(phash3))
}

func TestPixels(t *testing.T) {
	rgba := decodeFile(t, "rgba.png")
	allExpected := []*Pixel{
		{Red: 255},
		{Green: 255},
		{Blue: 255},
		{Opacity: 127},
	}
	leftExpected := []*Pixel{
		allExpected[0],
		allExpected[2],
	}
	rightExpected := []*Pixel{
		allExpected[1],
		allExpected[3],
	}
	runPixelTests := func() {
		px1, err := rgba.Pixels(rgba.Rect())
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(px1, allExpected) {
			t.Errorf("expecting pixels %v, got %v instead", allExpected, px1)
		}
		px2, err := rgba.Pixels(Rect{Width: 1, Height: 2})
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(px2, leftExpected) {
			t.Errorf("expecting left pixels %v, got %v instead", leftExpected, px2)
		}
		px3, err := rgba.Pixels(Rect{X: 1, Width: 1, Height: 2})
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(px3, rightExpected) {
			t.Errorf("expecting right pixels %v, got %v instead", rightExpected, px3)
		}
	}
	// First test with the image as it's loaded
	runPixelTests()
	// Change the green pixel to blue and test again
	if err := rgba.SetPixel(1, 0, &Pixel{Blue: 255}); err != nil {
		t.Fatal(err)
	}
	allExpected[1].Green = 0
	allExpected[1].Blue = 255
	runPixelTests()
	// Encode the image, decode it and check again
	var buf bytes.Buffer
	if err := rgba.Encode(&buf, nil); err != nil {
		t.Fatal(err)
	}
	var err error
	rgba, err = DecodeData(buf.Bytes())
	if err != nil {
		t.Fatal(err)
	}
	runPixelTests()
}

func BenchmarkRefUnref(b *testing.B) {
	im := decodeFile(b, "wizard.png")
	img := im.image
	b.ResetTimer()
	for ii := 0; ii < b.N; ii++ {
		refImage(img)
		unrefImage(img)
	}
}

func BenchmarkResizeAnimated(b *testing.B) {
	im := decodeFile(b, "Newtons_cradle_animation_book_2.gif")
	b.ResetTimer()
	for ii := 0; ii < b.N; ii++ {
		resized, err := im.Resize(240, 180, FQuadratic)
		if err != nil {
			b.Fatal(err)
		}
		resized.Dispose()
	}
}

func BenchmarkMinifyAnimated(b *testing.B) {
	im := decodeFile(b, "Newtons_cradle_animation_book_2.gif")
	coalesced, err := im.Coalesce()
	if err != nil {
		b.Fatal(err)
	}
	b.ResetTimer()
	for ii := 0; ii < b.N; ii++ {
		_, err = coalesced.Minify()
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkGifEncode(b *testing.B) {
	im := decodeFile(b, "Newtons_cradle_animation_book_2.gif")
	coalesced, err := im.Coalesce()
	if err != nil {
		b.Fatal(err)
	}
	b.ResetTimer()
	for ii := 0; ii < b.N; ii++ {
		_, err = coalesced.GifEncode()
		if err != nil {
			b.Fatal(err)
		}
	}
}
