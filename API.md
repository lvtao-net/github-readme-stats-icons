# API 文档

## 基础URL

```
http://localhost:8080
```

## 接口列表

### 1. GitHub Stats Card

获取GitHub用户的统计数据卡片。

**端点：** `GET /api`

**参数：**

| 参数 | 类型 | 必填 | 默认值 | 说明 |
|------|------|------|--------|------|
| username | string | 是 | - | GitHub用户名 |
| hide | string | 否 | - | 隐藏指定统计项，逗号分隔（stars,commits,prs,issues,contribs） |
| show | string | 否 | - | 显示额外统计项（reviews,discussions_started,discussions_answered,prs_merged,prs_merged_percentage） |
| show_icons | boolean | 否 | false | 显示图标 |
| hide_rank | boolean | 否 | false | 隐藏等级 |
| include_all_commits | boolean | 否 | false | 统计所有提交 |
| theme | string | 否 | default | 主题名称 |
| title_color | hex | 否 | - | 标题颜色 |
| text_color | hex | 否 | - | 文字颜色 |
| icon_color | hex | 否 | - | 图标颜色 |
| bg_color | hex | 否 | - | 背景颜色 |
| border_color | hex | 否 | - | 边框颜色 |
| ring_color | hex | 否 | - | 等级圆环颜色 |
| hide_border | boolean | 否 | false | 隐藏边框 |
| border_radius | number | 否 | 4.5 | 边框圆角 |
| card_width | number | 否 | 495 | 卡片宽度 |
| custom_title | string | 否 | - | 自定义标题 |
| number_format | string | 否 | short | 数字格式（short/long） |
| number_precision | number | 否 | - | 小数位数（0-2） |
| commits_year | number | 否 | - | 指定年份的提交数 |
| rank_icon | string | 否 | default | 等级图标（default/github/percentile） |
| text_bold | boolean | 否 | true | 文字加粗 |
| disable_animations | boolean | 否 | false | 禁用动画 |
| cache_seconds | number | 否 | 21600 | 缓存时间（秒） |
| token | string | 否 | - | GitHub Token（临时） |

**示例：**

```
GET /api?username=anuraghazra&show_icons=true&theme=radical
```

**响应：** SVG 图片

---

### 2. Top Languages Card

获取用户最常用的编程语言卡片。

**端点：** `GET /api/top-langs`

**参数：**

| 参数 | 类型 | 必填 | 默认值 | 说明 |
|------|------|------|--------|------|
| username | string | 是 | - | GitHub用户名 |
| layout | string | 否 | normal | 布局（normal/compact/donut/donut-vertical/pie） |
| langs_count | number | 否 | 5 | 显示语言数量（1-20） |
| hide | string | 否 | - | 隐藏指定语言，逗号分隔 |
| exclude_repo | string | 否 | - | 排除指定仓库，逗号分隔 |
| hide_title | boolean | 否 | false | 隐藏标题 |
| hide_progress | boolean | 否 | false | 隐藏进度条 |
| hide_border | boolean | 否 | false | 隐藏边框 |
| border_radius | number | 否 | 4.5 | 边框圆角 |
| card_width | number | 否 | 300 | 卡片宽度 |
| custom_title | string | 否 | - | 自定义标题 |
| theme | string | 否 | default | 主题名称 |
| title_color | hex | 否 | - | 标题颜色 |
| text_color | hex | 否 | - | 文字颜色 |
| bg_color | hex | 否 | - | 背景颜色 |
| stats_format | string | 否 | percentages | 统计格式（percentages/bytes） |
| size_weight | number | 否 | 1 | 代码量权重 |
| count_weight | number | 否 | 0 | 仓库数权重 |
| disable_animations | boolean | 否 | false | 禁用动画 |
| cache_seconds | number | 否 | 21600 | 缓存时间（秒） |
| token | string | 否 | - | GitHub Token（临时） |

**示例：**

```
GET /api/top-langs?username=anuraghazra&layout=compact&langs_count=8
```

**响应：** SVG 图片

---

### 3. Repo Pin Card

获取指定仓库的置顶卡片。

**端点：** `GET /api/pin`

**参数：**

| 参数 | 类型 | 必填 | 默认值 | 说明 |
|------|------|------|--------|------|
| username | string | 是 | - | GitHub用户名 |
| repo | string | 是 | - | 仓库名 |
| show_owner | boolean | 否 | false | 显示仓库所有者 |
| theme | string | 否 | default | 主题名称 |
| title_color | hex | 否 | - | 标题颜色 |
| text_color | hex | 否 | - | 文字颜色 |
| bg_color | hex | 否 | - | 背景颜色 |
| border_color | hex | 否 | - | 边框颜色 |
| hide_border | boolean | 否 | false | 隐藏边框 |
| border_radius | number | 否 | 4.5 | 边框圆角 |
| card_width | number | 否 | 400 | 卡片宽度 |
| custom_title | string | 否 | - | 自定义标题 |
| description_lines_count | number | 否 | - | 描述行数（1-3） |
| disable_animations | boolean | 否 | false | 禁用动画 |
| cache_seconds | number | 否 | 21600 | 缓存时间（秒） |
| token | string | 否 | - | GitHub Token（临时） |

**示例：**

```
GET /api/pin?username=anuraghazra&repo=github-readme-stats&show_owner=true
```

**响应：** SVG 图片

---

### 4. Gist Pin Card

获取指定Gist的置顶卡片。

**端点：** `GET /api/gist`

**参数：**

| 参数 | 类型 | 必填 | 默认值 | 说明 |
|------|------|------|--------|------|
| id | string | 是 | - | Gist ID |
| show_owner | boolean | 否 | false | 显示Gist所有者 |
| theme | string | 否 | default | 主题名称 |
| title_color | hex | 否 | - | 标题颜色 |
| text_color | hex | 否 | - | 文字颜色 |
| bg_color | hex | 否 | - | 背景颜色 |
| border_color | hex | 否 | - | 边框颜色 |
| hide_border | boolean | 否 | false | 隐藏边框 |
| border_radius | number | 否 | 4.5 | 边框圆角 |
| card_width | number | 否 | 400 | 卡片宽度 |
| custom_title | string | 否 | - | 自定义标题 |
| disable_animations | boolean | 否 | false | 禁用动画 |
| cache_seconds | number | 否 | 21600 | 缓存时间（秒） |
| token | string | 否 | - | GitHub Token（临时） |

**示例：**

```
GET /api/gist?id=bbfce31e0217a3689c8d961a356cb10d&show_owner=true
```

**响应：** SVG 图片

---

### 5. WakaTime Stats Card

获取WakaTime统计卡片。

**端点：** `GET /api/wakatime`

**参数：**

| 参数 | 类型 | 必填 | 默认值 | 说明 |
|------|------|------|--------|------|
| username | string | 是 | - | WakaTime用户名 |
| layout | string | 否 | default | 布局（default/compact） |
| langs_count | number | 否 | - | 显示语言数量 |
| hide | string | 否 | - | 隐藏指定语言 |
| hide_title | boolean | 否 | false | 隐藏标题 |
| hide_progress | boolean | 否 | false | 隐藏进度条 |
| theme | string | 否 | default | 主题名称 |
| title_color | hex | 否 | - | 标题颜色 |
| text_color | hex | 否 | - | 文字颜色 |
| bg_color | hex | 否 | - | 背景颜色 |
| border_color | hex | 否 | - | 边框颜色 |
| hide_border | boolean | 否 | false | 隐藏边框 |
| border_radius | number | 否 | 4.5 | 边框圆角 |
| card_width | number | 否 | 495 | 卡片宽度 |
| line_height | number | 否 | 25 | 行高 |
| custom_title | string | 否 | - | 自定义标题 |
| disable_animations | boolean | 否 | false | 禁用动画 |
| cache_seconds | number | 否 | 21600 | 缓存时间（秒） |

**示例：**

```
GET /api/wakatime?username=ffflabs&layout=compact
```

**响应：** SVG 图片

---

### 6. Skill Icons

获取技能图标组合图片。

**端点：** `GET /api/icons`

**参数：**

| 参数 | 类型 | 必填 | 默认值 | 说明 |
|------|------|------|--------|------|
| i | string | 是 | - | 图标列表，逗号分隔 |
| theme | string | 否 | dark | 主题（dark/light） |
| perline | number | 否 | 15 | 每行图标数量（1-50） |

**示例：**

```
GET /api/icons?i=js,html,css,react,vue,go&theme=dark&perline=6
```

**响应：** SVG 图片

**支持的图标：**

<details>
<summary>点击查看完整图标列表</summary>

**编程语言：**
`js`, `ts`, `java`, `py`, `go`, `rust`, `cpp`, `c`, `cs`, `php`, `rb`, `swift`, `kt`, `scala`, `r`, `perl`, `lua`, `dart`, `elixir`, `haskell`, `clojure`, `erlang`, `ocaml`, `nim`, `zig`, `crystal`, `groovy`, `vb`, `fsharp`

**Web技术：**
`html`, `css`, `sass`, `less`, `tailwind`, `bootstrap`, `materialui`, `chakra`, `antd`, `react`, `vue`, `angular`, `svelte`, `solidjs`, `preact`, `alpine`, `htmx`, `jquery`, `nextjs`, `nuxt`, `gatsby`, `remix`, `astro`, `webpack`, `vite`, `rollup`, `esbuild`, `parcel`, `gulp`, `babel`, `postcss`, `nodejs`, `deno`, `bun`, `express`, `koa`, `fastify`, `nestjs`, `graphql`, `rest`, `trpc`, `grpc`, `websocket`, `socketio`, `apollo`

**数据库：**
`mysql`, `postgres`, `mongodb`, `redis`, `sqlite`, `mariadb`, `cassandra`, `couchdb`, `dynamodb`, `firebase`, `supabase`, `prisma`, `sequelize`, `typeorm`, `mongoose`, `elasticsearch`, `neo4j`

**云服务：**
`aws`, `azure`, `gcp`, `vercel`, `netlify`, `heroku`, `digitalocean`, `linode`, `docker`, `kubernetes`, `terraform`, `ansible`, `jenkins`, `githubactions`, `gitlab`, `circleci`, `travis`

**工具：**
`git`, `github`, `gitlab`, `bitbucket`, `vscode`, `vim`, `neovim`, `sublime`, `atom`, `webstorm`, `intellij`, `pycharm`, `goland`, `phpstorm`, `rider`, `androidstudio`, `xcode`, `eclipse`

**操作系统：**
`linux`, `ubuntu`, `debian`, `fedora`, `arch`, `centos`, `redhat`, `opensuse`, `windows`, `macos`, `freebsd`, `kali`

**框架：**
`spring`, `django`, `flask`, `fastapi`, `rails`, `laravel`, `symfony`, `nestjs`, `gin`, `flutter`, `reactnative`, `electron`, `unity`, `unreal`, `godot`

</details>

---

### 7. Health Check

健康检查端点。

**端点：** `GET /health`

**响应：**

```json
{
  "status": "ok"
}
```

---

## 主题列表

| 主题名称 | 描述 |
|----------|------|
| `default` | 默认主题 |
| `dark` | 深色主题 |
| `radical` | 激进主题 |
| `merko` | 墨绿主题 |
| `gruvbox` | Gruvbox配色 |
| `tokyonight` | 东京夜景 |
| `onedark` | One Dark |
| `cobalt` | 钴蓝主题 |
| `synthwave` | 合成波 |
| `highcontrast` | 高对比度 |
| `dracula` | 德古拉 |
| `prussian` | 普鲁士蓝 |
| `monokai` | Monokai |
| `vue` | Vue风格 |
| `vue-dark` | Vue深色 |
| `github-dark` | GitHub深色 |
| `github-dark-blue` | GitHub深蓝 |
| `transparent` | 透明背景 |

---

## 颜色格式

所有颜色参数支持以下格式：

- 6位十六进制：`2f80ed`
- 3位十六进制：`f00`
- 带#号：`#2f80ed`

背景颜色支持渐变：

```
bg_color=45,ff0000,0000ff  # 45度角，红到蓝渐变
```

---

## 错误响应

| 状态码 | 说明 |
|--------|------|
| 400 | 缺少必填参数 |
| 403 | 用户/资源不在白名单中 |
| 404 | 用户/仓库/Gist不存在 |
| 500 | 服务器内部错误 |

---

## 缓存

所有API响应都包含缓存头：

```
Cache-Control: max-age=21600
```

可以通过 `cache_seconds` 参数覆盖默认缓存时间（21600-86400秒）。
