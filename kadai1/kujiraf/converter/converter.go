package converter

import (
	"fmt"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"
	"strings"
)

// supported extensions
const (
	extJpeg = ".jpeg"
	extPng  = ".png"
	extGif  = ".gif"
)

// Converter is image convertor.
// Src: Source directory. Required value.
// Dst: Destination directory.
// From: Extension before convert.
// To: Extension after converted.
// IsDebug: Converter debug flag.
type Converter struct {
	Src     string
	Dst     string
	From    string
	To      string
	IsDebug bool
}

// Validate validates flags
func (c *Converter) Validate() error {
	c.debugf("Flags : %+v\n", c)

	// ディレクトリにアクセス可能か
	f, err := os.Stat(c.Src)
	if err != nil {
		return fmt.Errorf("%s failed to get directory Error:%s", c.Src, err.Error())
	}
	if !f.IsDir() {
		return fmt.Errorf("%s is not directory", c.Src)
	}

	// -from, -toはサポート対象のものを指定しているか
	if ok := isSupported(&c.From); !ok {
		return fmt.Errorf("from ext %s is not supported", c.From)
	}
	if ok := isSupported(&c.To); !ok {
		return fmt.Errorf("to ext %s is not supported", c.To)
	}

	// -fromと-toが同じ値ではないか
	if c.From == c.To {
		return fmt.Errorf("-from and -to are same. -from %s, -to %s", c.From, c.To)
	}

	return nil
}

func isSupported(ext *string) bool {
	if !strings.Contains(*ext, ".") {
		*ext = "." + *ext
	}
	switch *ext {
	case ".jpeg", ".jpg":
		*ext = extJpeg
		return true
	case extPng, extGif:
		return true
	default:
		return false
	}
}

// DoConvert converts image's extension from c.From to c.To.
func (c *Converter) DoConvert() (err error) {

	// 出力先パスを絶対パスに変える
	if !filepath.IsAbs(c.Dst) {
		abs, err := filepath.Abs(c.Dst)
		if err != nil {
			return err
		}
		c.Dst = abs
	}
	c.debugf("output root path : %s\n", c.Dst)

	// 処理対象のディレクトリに移動する。処理が終われば元の場所に戻る。
	prevDir, err := filepath.Abs(".")
	if err != nil {
		return err
	}
	c.debugf("current dir : %s\n", prevDir)
	err = os.Chdir(c.Src)
	if err != nil {
		return err
	}
	defer func() { setNewError(os.Chdir(prevDir), &err) }()

	// 処理対象のディレクトリ名取得
	workDir, err := os.Getwd()
	if err != nil {
		return err
	}
	srcRoot := filepath.Base(workDir)
	c.debugf("src dir name : %s\n", srcRoot)

	// 再帰的に処理を実行する
	err = filepath.Walk(".",
		func(path string, info os.FileInfo, err error) error {
			if filepath.Ext(path) == c.From {
				c.debugf("found. %s\n", path)

				// 対象のファイルに対して処理を実行する
				err := c.convert(path, srcRoot)
				if err != nil {
					return err
				}

			}
			return nil
		})
	if err != nil {
		return err
	}
	return nil
}

func (c *Converter) convert(inputfile string, root string) error {
	c.debugf("target file path=%v\n", inputfile)

	// 出力先ディレクトリの作成
	dst := filepath.Dir(filepath.Join(c.Dst, root, inputfile))
	c.debugf("output path=%v\n", dst)
	err := os.MkdirAll(dst, os.ModeDir)
	if err != nil {
		return err
	}

	// ファイルのデコード
	img, err := c.decode(inputfile)
	if err != nil {
		return err
	}

	// 出力ファイル名取得
	extlen := len(inputfile) - len(filepath.Ext(inputfile))
	filename := filepath.Base(inputfile[:extlen]) + c.To
	outputfile := filepath.Join(dst, filename)

	// ファイルを出力先にコピー
	err = c.encode(outputfile, img)
	if err != nil {
		return err
	}

	fmt.Printf("[INFO]conversion complete. converted file from %s to %s\n", inputfile, outputfile)
	return nil
}

func (c Converter) decode(input string) (img image.Image, err error) {
	in, err := os.Open(input)
	if err != nil {
		return nil, err
	}
	defer func() { setNewError(in.Close(), &err) }()

	switch c.From {
	case extPng:
		c.debugf("decode %s file %s\n", extPng, input)
		img, err = png.Decode(in)
		return
	case extGif:
		c.debugf("decode %s file %s\n", extGif, input)
		img, err = gif.Decode(in)
		return
	default:
		c.debugf("decode %s file %s\n", extJpeg, input)
		img, err = jpeg.Decode(in)
		return
	}
}

func (c Converter) encode(output string, m image.Image) (err error) {
	newfile, err := os.Create(output)
	if err != nil {
		return err
	}
	defer func() { setNewError(newfile.Close(), &err) }()

	switch c.To {
	case extJpeg:
		c.debugf("encode %s file and output to %s\n", extJpeg, newfile.Name())
		options := &jpeg.Options{Quality: 100}
		err = jpeg.Encode(newfile, m, options)
		return
	case extGif:
		c.debugf("encode %s file and output to %s\n", extGif, newfile.Name())
		options := &gif.Options{NumColors: 256}
		err = gif.Encode(newfile, m, options)
		return
	default:
		c.debugf("encode %s file and output to %s\n", extPng, newfile.Name())
		err = png.Encode(newfile, m)
		return
	}
}

func setNewError(newErr error, currentErr *error) {
	if *currentErr == nil {
		currentErr = &newErr
	} else if currentErr != nil && newErr != nil {
		fmt.Println("[ERROR]", newErr.Error())
	}
}

func (c Converter) debugf(format string, a ...interface{}) {
	c.debug(func() string {
		return fmt.Sprintf(format, a...)
	})
}

func (c Converter) debug(msg func() string) {
	if c.IsDebug {
		fmt.Print("[Debug]", msg())
	}
}
