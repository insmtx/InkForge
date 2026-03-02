# InkForge

高性能 Markdown 转图片服务。支持数学公式、代码高亮、流程图。

## 功能特性

- **Markdown 转图片** - 支持 PNG、JPEG、WebP 格式
- **LaTeX 数学公式** - `$E=mc^2$`、`$$\int_0^\infty$$` 等
- **代码高亮** - Python、JavaScript、Go、Bash 等
- **Mermaid 流程图** - 流程图、时序图等
- **主题模式** - 浅色和深色模式
- **高清输出** - 支持 Retina 画质

## 快速开始

### Docker 运行（推荐）

```bash
# 运行容器
docker run -d -p 8080:8080 inkforge

# 打开浏览器
http://localhost:8080
```

### 源码运行

```bash
go build -o inkforge ./cmd/inkforge/
./inkforge
```

## 使用方法

### API 接口

```
POST /api/v1/markdown2image
```

直接返回图片数据。

### cURL 示例

**最简单用法：**
```bash
curl -X POST http://localhost:8080/api/v1/markdown2image \
  -H "Content-Type: application/json" \
  -d '{"content": "# 你好世界"}' \
  -o image.jpg
```

**完整示例（数学 + 代码 + 流程图）：**
```bash
curl -X POST http://localhost:8080/api/v1/markdown2image \
  -H "Content-Type: application/json" \
  -d '{
    "content": "# 数学示例\n\n行内公式：$E=mc^2$\n\n块级公式：\n\n$$\\frac{-b \\pm \\sqrt{b^2-4ac}}{2a}$$\n\n## 代码\n\n```python\ndef fib(n):\n    if n <= 1:\n        return n\n    return fib(n-1) + fib(n-2)\n```\n\n## 流程图\n\n```mermaid\ngraph TD\n    A[开始] --> B{判断}\n    B -->|是| C[处理]\n    B -->|否| D[结束]\n```",
    "theme": "dark",
    "width": 800
  }' \
  -o output.png
```

**Python 调用示例：**
```python
import requests

response = requests.post(
    "http://localhost:8080/api/v1/markdown2image",
    json={
        "content": "# 你好\n\n**粗体**和*斜体*",
        "theme": "light",
        "width": 600
    }
)

with open("output.jpg", "wb") as f:
    f.write(response.content)

print(f"图片大小: {len(response.content)} bytes")
```

## 参数说明

| 参数 | 默认值 | 说明 |
|------|--------|------|
| content | （必填） | Markdown 内容 |
| theme | "light" | 主题："light" 或 "dark" |
| image_format | "jpg" | 图片格式："jpg"、"png"、"webp" |
| width | 1200 | 图片宽度（像素） |
| height | 800 | 图片高度（像素） |
| scale | 2.0 | 缩放比例（2.0 = 2倍清晰度） |
| quality | 90 | JPEG 质量（1-100） |

## 示例

### 数学公式

```markdown
行内公式：$E=mc^2$

块级公式：
$$
\int_0^\infty e^{-x^2} dx = \frac{\sqrt{\pi}}{2}
$$
```

### 代码块

````markdown
```python
def hello():
    print("你好")
```
````

### 流程图

````markdown
```mermaid
graph TD
    A --> B
```
````

## 健康检查

```bash
curl http://localhost:8080/api/v1/health
```

## 许可

MIT License
