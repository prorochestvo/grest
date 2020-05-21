package helper

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
)

var HtmlIcon string = `data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAABgAAAAYCAYAAADgdz34AAAABmJLR0QA/wD/AP+gvaeTAAAACXBIWXMAAA7EAAAOxAGVKw4bAAAAB3RJTUUH5AMNAyouRTAuhwAAAr5JREFUSMet1U2IllUUB/DfnU+ZxhLlHaIsoS8pZYREGmzRRoIIhGBw80SbPnDhIgIh00XuBF0plQq1CO6qwD4IpA+oVSC1iD6IWmUkWVoRM4rO6HXReYbbw+vwCp7NfZ577vmfc885/3OTAaVkCStRMJcaZRC7NAAwDOMFPIOryDiGK6lZ3n6oAtHvOwA243XMYCtew8OpWd52yQFSyaZL1usYDZdsBm/2Ce6t0A3XwZSsV7LpNjutg/vxNd6vIhjFPnyJjX0cbAzd3pKNVPvvBdYDtYNxzGEtS2m5Ha8OUMf9uKOqxZ2BNVY7OIsFjJVsdexNGVym4tarsSKwzi45SI0/8BtuwyNhdP4GHPwV60xgnAlMQ1VBD4f35yOnv+PDAcA/wJmweS4wDi8VvtPz30bxDuIALmAXnkav4k3Bn3gbRwL0FbyMH1Jjw/+IVrXYWnyE6UjRCZzEV/gHi2374lZswjZsxz34BttT43SLmbqkKtkTeBcTobqMeVwKFre1G48z47E3j9nUOFkRVCrZqgqs4LEg1kSffP8S67o+uot4Fp9X3TmXSnYKWwbslg2R1u8GPH9qJObK47gSm7fgyerqqhS07TiHyY7+UtRvvkrjx+k6E/Q07ups/40H4/tHrOrof02Nu5edpiVbX7I3sKaf3+r7ah/9mpIdLdn6GjNV3bMt2nIS5zAarFQxu+3v7ztB/Bsp6kX6nkqNT0tmKMDvDfBxHEqNHmZvYFTMpsYUDgXGiZLdl5r/RkXCzoj8aGrsDqPPsGeZNLWyB5/ETNuN44G1s2RpKH4ejcMvVaQrMS52hG60Ah2LdQcOdB6pF2PdipUtI9fh59RYrFkYUb0TQRyJMXw5htlk6NpzbWCL+Ckwx0aCgV9E63XBW6P5eN1a2dd5s3VsMx7CxRRXm8BCaiy4CRLP7SguXAN8xebnDAQRCAAAAABJRU5ErkJggg==`

func HtmlDocument(body string, style ...string) string {
	const cssMarginPadding string = `.m-l-none{margin-left:0}.m-l-xxs{margin-left:2px}.m-l-xs{margin-left:5px}.m-l-sm{margin-left:10px}.m-l{margin-left:15px}.m-l-md{margin-left:20px}.m-l-lg{margin-left:30px}.m-l-xl{margin-left:40px}.m-l-n-xxs{margin-left:-1px}.m-l-n-xs{margin-left:-5px}.m-l-n-sm{margin-left:-10px}.m-l-n{margin-left:-15px}.m-l-n-md{margin-left:-20px}.m-l-n-lg{margin-left:-30px}.m-l-n-xl{margin-left:-40px}.m-t-none{margin-top:0}.m-t-xxs{margin-top:1px}.m-t-xs{margin-top:5px}.m-t-sm{margin-top:10px}.m-t{margin-top:15px}.m-t-md{margin-top:20px}.m-t-lg{margin-top:30px}.m-t-xl{margin-top:40px}.m-t-n-xxs{margin-top:-1px}.m-t-n-xs{margin-top:-5px}.m-t-n-sm{margin-top:-10px}.m-t-n{margin-top:-15px}.m-t-n-md{margin-top:-20px}.m-t-n-lg{margin-top:-30px}.m-t-n-xl{margin-top:-40px}.m-r-none{margin-right:0}.m-r-xxs{margin-right:1px}.m-r-xs{margin-right:5px}.m-r-sm{margin-right:10px}.m-r{margin-right:15px}.m-r-md{margin-right:20px}.m-r-lg{margin-right:30px}.m-r-xl{margin-right:40px}.m-r-n-xxs{margin-right:-1px}.m-r-n-xs{margin-right:-5px}.m-r-n-sm{margin-right:-10px}.m-r-n{margin-right:-15px}.m-r-n-md{margin-right:-20px}.m-r-n-lg{margin-right:-30px}.m-r-n-xl{margin-right:-40px}.m-b-none{margin-bottom:0}.m-b-xxs{margin-bottom:1px}.m-b-xs{margin-bottom:5px}.m-b-sm{margin-bottom:10px}.m-b{margin-bottom:15px}.m-b-md{margin-bottom:20px}.m-b-lg{margin-bottom:30px}.m-b-xl{margin-bottom:40px}.m-b-n-xxs{margin-bottom:-1px}.m-b-n-xs{margin-bottom:-5px}.m-b-n-sm{margin-bottom:-10px}.m-b-n{margin-bottom:-15px}.m-b-n-md{margin-bottom:-20px}.m-b-n-lg{margin-bottom:-30px}.m-b-n-xl{margin-bottom:-40px}.no-padding{padding:0!important}.no-borders{border:none!important}.no-margins{margin:0!important}`
	const cssFontStyle string = `small,.small{font-size:77%;}.font-bold{font-weight:600}.font-normal{font-weight:400}.font-italic{font-style:italic}.text-uppercase{text-transform:uppercase}.text-lowercase{text-transform:lowercase}`
	const cssFullSize string = `.full-width{width:100%!important}.full-height{height:100%!important}`
	css := make([]string, 0)
	css = append(css, `
body {
  font-family: "open sans";
  font-size: 13px;
  overflow-x: hidden;
  padding: 0;
  background-color: #f5f5f5;
  color: #676a6c;
}
a { cursor: pointer; color: #676a6c; text-decoration: none; }
a:hover,
a:focus { color: #1ab394; }
a.link:after { content: "\27A2"; }
.text-monospaced { font-family: Courier New,Monaco !important; }
ul { padding: 0; margin: 0; list-style-type: none; }
ul.hor li { display:inline; }
`)
	css = append(css, cssMarginPadding)
	css = append(css, cssFontStyle)
	css = append(css, cssFullSize)
	for _, s := range style {
		if _, err := os.Stat(s); !os.IsNotExist(err) {
			if bytes, err := ioutil.ReadFile(s); err == nil && len(bytes) > 0 {
				s = string(bytes)
			}
		}
		css = append(css, s)
	}
	return fmt.Sprintf(`
<!DOCTYPE html>
<html>
	<head>
		<meta charset="utf-8">
		<title>API Documentation</title>
		<link rel="icon" href="%s">
	</head>
	%s
	<body>%s</body>
</html>`, HtmlIcon, fmt.Sprintf("<style>%s</style>", strings.Join(css, "\n")), body)
}
