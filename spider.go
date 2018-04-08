package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"io"
	"net/http"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

//定义spider数据类型
type Spider struct {
	url    string
	header map[string]string
}

type Book struct {
	Name  string
	Score int
	Star  int
}
type Bookslice []Book

var Url_array []string

func (b Bookslice) Len() int { return len(b) }

func (b Bookslice) Swap(i, j int) { b[i], b[j] = b[j], b[i] }

func (b Bookslice) Less(i, j int) bool { return b[i].Score < b[j].Score }

var File_path []string

//定义 Spider get的方法
func (sp Spider) get_html_header() string {
	client := &http.Client{}
	req, err := http.NewRequest("GET", sp.url, nil)
	if err != nil {
	}
	for key, value := range sp.header {
		req.Header.Add(key, value)
	}
	resp, err := client.Do(req)
	if err != nil {
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
	}
	return string(body)

}

//在tag页面抓取信息
func spider_at_tag(url_tag string) {

	header := map[string]string{
		"Host":                      "movie.douban.com",
		"Connection":                "keep-alive",
		"Cache-Control":             "max-age=0",
		"Upgrade-Insecure-Requests": "1",
		"User-Agent":                "Mozilla/5.0 (Windows NT 6.1; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/53.0.2785.143 Safari/537.36",
		"Accept":                    "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8",
		"Referer":                   "https://movie.douban.com/top250",
	}

	//获取tag
	spider_tag := &Spider{url_tag, header}
	html_tag := spider_tag.get_html_header()
	pattern_tag := `<h1>豆瓣图书标签: (.*?)</h1>`
	rp_tag := regexp.MustCompile(pattern_tag)
	tag := rp_tag.FindAllStringSubmatch(html_tag, -1)

	//创建tag文件
	tag_file := "/home/ec2-user/workhome/spider/" + string(tag[0][1])
	File_path = append(File_path, tag_file)
	f, err := os.Create(tag_file)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	//循环每页解析并把结果写入file
	for i := 0; i <= 50; i++ {
		fmt.Println("正在抓取第" + strconv.Itoa(i) + "页......")
		url := url_tag + "?start=" + strconv.Itoa(i*20) + "&type=T"
		//fmt.Println(url)
		spider := &Spider{url, header}
		html := spider.get_html_header()

		//评价人数
		pattern2 := `\((.*?)人评价\)`
		rp2 := regexp.MustCompile(pattern2)
		find_txt2 := rp2.FindAllStringSubmatch(html, -1)

		//评分
		pattern3 := `<span class="rating_nums">(.*?)</span>`
		rp3 := regexp.MustCompile(pattern3)
		find_txt3 := rp3.FindAllStringSubmatch(html, -1)

		//图书名称
		pattern4 := `<a href="https://book.douban.com/subject/[0-9]*/" title="(.*?)"`
		rp4 := regexp.MustCompile(pattern4)
		find_txt4 := rp4.FindAllStringSubmatch(html, -1)

		// 写入UTF-8 BOM
		//f.WriteString("\xEF\xBB\xBF")
		//  打印全部数据和写入文件
		for i := 0; i < len(find_txt2); i++ {
			//fmt.Printf("%s   %s   %s\n", find_txt4[i][1], find_txt3[i][1], find_txt2[i][1])
			f.WriteString(find_txt4[i][1] + "\t" + find_txt3[i][1] + "\t" + find_txt2[i][1] + "\r\n")
		}
	}
}

func spider_all(url string) {

	header := map[string]string{
		"Host":                      "movie.douban.com",
		"Connection":                "keep-alive",
		"Cache-Control":             "max-age=0",
		"Upgrade-Insecure-Requests": "1",
		"User-Agent":                "Mozilla/5.0 (Windows NT 6.1; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/53.0.2785.143 Safari/537.36",
		"Accept":                    "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8",
		"Referer":                   "https://movie.douban.com/top250",
	}

	//获取tag的url
	spider_url := &Spider{url, header}
	html_url := spider_url.get_html_header()
	pattern_url := `<td><a href="(.*?)">`
	rp_url := regexp.MustCompile(pattern_url)
	tag_url := rp_url.FindAllStringSubmatch(html_url, -1)

	//tag url保存起来
	for i := 0; i < len(tag_url); i++ {
		full_url := "https://book.douban.com" + string(tag_url[i][1])
		Url_array = append(Url_array, full_url)
		spider_at_tag(full_url)
	}
}

func parse_date() {
	for i := 0; i < len(File_path); i++ {
		fmt.Println("处理文件：" + File_path[i])
		f, err := os.Open(File_path[i])
		if err != nil {
			panic(err)
		}
		defer f.Close()

		var book Bookslice
		var bb Book
		br := bufio.NewReader(f)
		for {
			a, _, c := br.ReadLine()
			if c == io.EOF {
				break
			}
			ba := strings.Split(string(a), "\t")
			bb.Name = ba[0]
			bb.Score, err =strconv.Atoi(ba[1])
			bb.Star, err = strconv.Atoi(ba[2])
			book = append(book, bb)
			//fmt.Println(string(a))
		}
		sort.Stable(book)
		for i := 0; i < 100; i++ {
			fmt.Println(book[i].Name + "\t" + string(book[i].Score) + "\t" + string(book[i].Star))
		}

	}
}

func main() {
	t1 := time.Now() // get current time
	spider_all("https://book.douban.com/tag/?view=cloud")
	elapsed := time.Since(t1)
	fmt.Println("爬虫结束,总共耗时: ", elapsed)
	fmt.Println("开始处理数据...")
	parse_date()
	fmt.Println("处理完成！")
}
