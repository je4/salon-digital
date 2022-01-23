package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/Rayleigh865/gopack"
	"github.com/disintegration/imaging"
	"github.com/pkg/errors"
	"image"
	"image/color"
	"image/draw"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"log"
	"os"
	"path/filepath"
)

const borderWidth = 3
const offsetX = 20
const offsetY = 20

func loadImage(filePath string) (image.Image, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, errors.Wrapf(err, "cannot open image %s", filePath)
	}
	defer f.Close()

	image, _, err := image.Decode(f)
	if err != nil {
		return nil, errors.Wrapf(err, "cannot decode image %s", filePath)
	}

	return image, nil
}

func Rect(x1, y1, x2, y2, thickness int, col color.Color, img *image.NRGBA) {

	for t := 0; t < thickness; t++ {
		// draw horizontal lines
		for x := x1; x <= x2; x++ {
			img.Set(x, y1+t, col)
			img.Set(x, y2-t, col)
		}
		// draw vertical lines
		for y := y1; y <= y2; y++ {
			img.Set(x1+t, y, col)
			img.Set(x2-t, y, col)
		}
	}
}

type VirtualImagePosition struct {
	X, Y          int
	Width, Height int
	Rotation      bool
}

type VirtualImages struct {
	OffsetX, OffsetY int
	Filename         string
	Image            map[string]VirtualImagePosition
}

func main() {
	var basedir = flag.String("basedir", ".", "base folder with html contents")
	var configfile = flag.String("cfg", "/etc/collage.toml", "configuration file")

	flag.Parse()

	var config = &CollageConfig{
		BaseDir: *basedir,
	}
	if err := LoadCollageConfig(*configfile, config); err != nil {
		log.Fatalf("cannot load config file %s: %v", *configfile, err)
	}

	p := gopack.NewPacker()
	// Add bin
	for page := 1; page <= 10; page++ {
		p.AddBin(gopack.NewBin(fmt.Sprintf("Page%02d", page), 595*2-offsetX, 842*2-offsetY))
	}

	var images = map[string]image.Image{}
	for name, filePath := range config.Image {
		image, err := loadImage(filepath.Join(config.BaseDir, filePath))
		if err != nil {
			log.Fatal(err)
		}
		bounds := image.Bounds()
		images[name] = image
		p.AddItem(gopack.NewItem(name, int64(bounds.Dx()+2*borderWidth), int64(bounds.Dy()+2*borderWidth)))
	}
	// Solve
	if err := p.Pack(); err != nil {
		log.Fatal(err)
	}

	//show results
	gopack.Display_packed(p.Bins)
	for _, bin := range p.Bins {
		if len(bin.Items) > 0 {
			vi := &VirtualImages{
				OffsetX:  offsetX,
				OffsetY:  offsetY,
				Filename: fmt.Sprintf("%s.png", bin.Name),
				Image:    map[string]VirtualImagePosition{},
			}
			img := imaging.New(int(bin.GetWidth())+offsetX, int(bin.GetHeight()+offsetY), color.Transparent)
			for _, item := range bin.Items {
				miniImg := images[item.Name]
				width := int(item.GetWidth() - 2*borderWidth)
				height := int(item.GetHeight() - 2*borderWidth)

				switch item.RotationType {
				case gopack.RotationType_WH:
					// nothing to do
				case gopack.RotationType_HW:
					// rotate 90°
					miniImg = imaging.Rotate(miniImg, 90, color.Transparent)
					mFilename := filepath.Join(config.OutputDir, fmt.Sprintf("%s.png", item.Name))
					if err := imaging.Save(miniImg, mFilename); err != nil {
						log.Fatalf("cannot store %s: %v", mFilename, err)
					}
					width = int(item.GetHeight() - 2*borderWidth)
					height = int(item.GetWidth() - 2*borderWidth)
				default:
					log.Fatalf("invalid rotation type for item %s: %v", item.Name, item.RotationType)
				}
				posX := int(item.Position[0] + borderWidth)
				posY := int(item.Position[1] + borderWidth)
				vip := VirtualImagePosition{
					X:        posX,
					Y:        posY,
					Width:    width,
					Height:   height,
					Rotation: item.RotationType == gopack.RotationType_HW,
				}
				draw.Draw(img, image.Rect(posX+offsetX, posY+offsetY,
					posX+width+offsetX, posY+height+offsetY),
					miniImg, image.Pt(0, 0), draw.Src)
				Rect(posX-2+offsetX, posY-2+offsetY,
					posX+width+1+offsetX, posY+height+1+offsetY, 1, color.Black, img)
				vi.Image[item.Name] = vip
			}
			filename := filepath.Join(config.OutputDir, fmt.Sprintf("%s.png", bin.Name))
			if err := imaging.Save(img, filename); err != nil {
				log.Fatalf("cannot store %s: %v", filename, err)
			}
			data, err := json.MarshalIndent(vi, "", "   ")
			if err != nil {
				log.Fatalf("cannot marshal json: %v", err)
			}
			filename = filepath.Join(config.OutputDir, fmt.Sprintf("%s.json", bin.Name))
			if err := os.WriteFile(filename, data, 0666); err != nil {
				log.Fatalf("cannot store file %s: %v", filename, err)
			}
		}
	}
}
