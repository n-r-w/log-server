package httprouter

import (
	g "github.com/maragudk/gomponents"
	c "github.com/maragudk/gomponents/components"
	. "github.com/maragudk/gomponents/html"
)

// Информация о странице навбара
type pageInfo struct {
	path string
	name string
}

// Группа страниц, объединенных одним пунктом в навбаре
type pageInfoGroup struct {
	pages []pageInfo
}

// Загружаемые css
var (
	stylesExternal = []string{
		"https://unpkg.com/tailwindcss@2.1.2/dist/base.min.css",
		"https://unpkg.com/tailwindcss@2.1.2/dist/components.min.css",
		"https://unpkg.com/@tailwindcss/typography@0.4.0/dist/typography.min.css",
		"https://unpkg.com/tailwindcss@2.1.2/dist/utilities.min.css",
		"https://unpkg.com/flowbite@1.4.4/dist/flowbite.min.css",
		"https://use.fontawesome.com/releases/v5.11.2/css/all.css",
		"https://fonts.googleapis.com/css2?family=Inter:wght@300;400;500;600;700&display=swap",
	}
)

var (
	// Описание всех страниц навбара
	navInfo = []pageInfoGroup{
		{
			pages: []pageInfo{
				{
					path: "/",
					name: "Просмотр",
				},
				{
					path: "/search",
					name: "Поиск",
				},
			},
		},
		{
			pages: []pageInfo{
				{
					path: "/stats",
					name: "Статистика",
				},
			},
		},
		{
			pages: []pageInfo{
				{
					path: "/admin",
					name: "Администрирование",
				},
			},
		},
		{
			pages: []pageInfo{
				{
					path: "/login",
					name: "Личный кабинет",
				},
			},
		},
	}
)

// Содержится ли такой путь в группе страниц, относящихся к одному пункту навбара
func (p *pageInfoGroup) contains(path string) bool {
	for _, page := range p.pages {
		if page.path == path {
			return true
		}
	}
	return false
}

// Поиск информации о пунктах навбара по пути
func getNavInfoByPath(path string) *pageInfo {
	for _, info := range navInfo {
		for _, p := range info.pages {
			if p.path == path {
				return &p
			}
		}
	}
	return nil
}

// Отрисовка страницы
func page(title, path string, body g.Node) g.Node {
	// стили
	var styles []g.Node
	for _, style := range stylesExternal {
		styles = append(styles, Link(Rel("stylesheet"), Href(style)))
	}

	// 	styles = append(styles, StyleAttr(`
	// .picker__date-display {
	//   background-color:blue;
	// }
	// .picker__weekday-display {
	//   background-color:red;
	// }
	// .picker__day--selected, .picker__day--selected:hover, .picker--focused .picker__day--selected {
	//   background-color:blue;
	// }
	// `))

	return c.HTML5(c.HTML5Props{
		Title:    title,
		Language: "ru",
		Head:     styles,
		Body: []g.Node{colorStyleAttr, Class("bg-gray-700 mb-3 max-w-7xl mx-0 px-0 sm:px-0 lg:px-0"),
			Div(navbar(path, navInfo),
				Div(Class("prose-sm px-5 py-0 mt-0"), body)),
		},
	})
}

// Рендер навбара
func navbar(currentPath string, navInfo []pageInfoGroup) g.Node {

	return Nav(Class("bg-gray-700 mb-3 max-w-7xl mx-0 px-0 sm:px-0 lg:px-0"),
		Div(Class("flex flex-wrap items-center space-x-0"),
			B(H3(Class("text-4xl py-0 px-5 h-12 "), StyleAttr("color:#91B4FF"),
				g.Text("Логгер"))),
			Div(Class("mt-0"), g.Group(g.Map(len(navInfo),
				func(i int) g.Node {
					pGroup := navInfo[i]
					return navbarLink(pGroup.pages[0].path, pGroup.pages[0].name, pGroup.contains(currentPath))
				}))),
		),
	)
}

// Рендер элемента навбара
func navbarLink(path, text string, active bool) g.Node {
	return A(Href(path), g.Text(text),
		c.Classes{
			"flex-wrap px-2 py-2 mb-1 rounded-md text-sm font-medium focus:outline-none " +
				"focus:text-white focus:bg-gray-700": true,
			"text-white bg-gray-900":                           active,
			"text-gray-300 hover:text-white hover:bg-gray-700": !active,
		},
	)
}

func (router *HTTPRouter) renderNotLoginGeneral() g.Node {
	body := Div(Class("text-gray-300 hover:text-white "),
		A(StyleAttr(`border-bottom: 1px solid grey; padding-bottom: 5px;`),
			Href("/login"), g.Text("Необходимо войти в личный кабинет")))
	return body
}

func (router *HTTPRouter) renderNotImplemeted() g.Node {
	body := Div(Class("text-white"),
		g.Text("Тут пока еще ничего нет"))
	return body
}
