package app

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"sync"
)

type ConfigFile struct {
	lock      sync.RWMutex
	fileNames []string
	data      map[string]string
	BlockMode bool
}

/**
 * 加载配置文件初始化函数
 *
 * @author Vckai.
 * @date   2014-04-03
 */
func LoadConfigFile(fileName string, moreFiles ...string) *ConfigFile {
	fileNames := make([]string, 1, len(moreFiles)+1)
	fileNames[0] = fileName
	if len(moreFiles) > 0 {
		fileNames = append(fileNames, moreFiles...)
	}
	c := newConfigFile(fileNames)

	for _, name := range fileNames {
		if err := c.loadFile(name); err != nil {
			fmt.Println("Read Config File Error: ", err)
		}
	}
	return c
}

func newConfigFile(fileNames []string) *ConfigFile {
	c := new(ConfigFile)
	c.fileNames = fileNames
	c.data = make(map[string]string)
	c.BlockMode = false

	return c
}

/**
 * 读取文件
 */
func (this *ConfigFile) loadFile(fileName string) (err error) {
	f, err := os.Open(fileName)
	if err != nil {
		return err
	}
	defer f.Close()

	return this.read(f)
}

/**
 * 读取文件写入内存
 */
func (this *ConfigFile) read(reader io.Reader) (err error) {
	buf := bufio.NewReader(reader)

	for {
		line, err := buf.ReadString('\n')
		line = strings.TrimSpace(line)
		lineLenth := len(line)
		if err != nil {
			if err != io.EOF {
				return err
			}

			if lineLenth == 0 {
				break
			}
		}

		switch {
		case lineLenth == 0:
			continue
		case line[0] == '#' || line[0] == ';':
			continue
		case line[0] == '[' && line[lineLenth-1] == ']':
			continue
		default:
			i := strings.IndexAny(line, "=")
			if i <= 0 {
				return errors.New(fmt.Sprintf("could not parse line: %s", line))
			}
			key := strings.TrimSpace(line[:i])
			val := strings.TrimSpace(line[i+1:])
			this.SetData(key, val)
		}
		if err == io.EOF {
			break
		}
	}
	return nil
}

/**
 * 设置配置项
 */
func (this *ConfigFile) SetData(key, val string) {
	if this.BlockMode {
		this.lock.Lock()
		defer this.lock.Unlock()
	}
	this.data[key] = val
}

/**
 * 获取配置
 */
func (this *ConfigFile) GetValue(key string) (string, error) {
	if this.BlockMode {
		this.lock.RLock()
		defer this.lock.RUnlock()
	}
	value, ok := this.data[key]
	if !ok || len(value) == 0 {
		return "", errors.New(fmt.Sprintf("Key '%s' Not Found", key))
	}
	return value, nil
}

/**
 * 获取bool类型的配置值
 */
func (this *ConfigFile) Bool(key string) (bool, error) {
	value, err := this.GetValue(key)
	if err != nil {
		return false, err
	}
	return strconv.ParseBool(value)
}

/**
 * 获取配置值, 并转换为浮点型
 */
func (this *ConfigFile) Float64(key string) (float64, error) {
	value, err := this.GetValue(key)
	if err != nil {
		return 0.0, err
	}
	return strconv.ParseFloat(value, 64)
}

/**
 * 获取配置值, 并转换为整型
 */
func (this *ConfigFile) Int(key string) (int, error) {
	value, err := this.GetValue(key)
	if err != nil {
		return 0, err
	}
	return strconv.Atoi(value)
}

/**
 * 获取配置值, 并转换为64位整型
 */
func (this *ConfigFile) Int64(key string) (int64, error) {
	value, err := this.GetValue(key)
	if err != nil {
		return 0, err
	}
	return strconv.ParseInt(value, 10, 64)
}

/**
 * 获取配置值, 字符串类型, 可传入默认值, 如果不存在则返回默认值
 */
func (this *ConfigFile) MustValue(key string, defaultValue ...string) string {
	value, err := this.GetValue(key)
	if err != nil && len(defaultValue) > 0 {
		return defaultValue[0]
	}
	return value
}

func (this *ConfigFile) MustBool(key string, defaultValue ...bool) bool {
	value, err := this.Bool(key)
	if err != nil && len(defaultValue) > 0 {
		return defaultValue[0]
	}
	return value
}
func (this *ConfigFile) MustFloat64(key string, defaultValue ...float64) float64 {
	value, err := this.Float64(key)
	if err != nil && len(defaultValue) > 0 {
		return defaultValue[0]
	}
	return value
}

func (this *ConfigFile) MustInt(key string, defaultValue ...int) int {
	value, err := this.Int(key)
	if err != nil && len(defaultValue) > 0 {
		return defaultValue[0]
	}
	return value
}

func (this *ConfigFile) MustInt64(key string, defaultValue ...int64) int64 {
	value, err := this.Int64(key)
	if err != nil && len(defaultValue) > 0 {
		return defaultValue[0]
	}
	return value
}
