import type { ComposeTemplate } from '@/types'

export const composeTemplates: ComposeTemplate[] = [
  {
    name: 'WordPress',
    description: 'WordPress blog with MySQL database',
    category: 'CMS',
    content: `version: "3.8"
services:
  wordpress:
    image: wordpress:latest
    ports:
      - "80:80"
    environment:
      WORDPRESS_DB_HOST: db
      WORDPRESS_DB_USER: wordpress
      WORDPRESS_DB_PASSWORD: wordpress
      WORDPRESS_DB_NAME: wordpress
    volumes:
      - wordpress_data:/var/www/html
    depends_on:
      - db
  db:
    image: mysql:8.0
    environment:
      MYSQL_ROOT_PASSWORD: rootpassword
      MYSQL_DATABASE: wordpress
      MYSQL_USER: wordpress
      MYSQL_PASSWORD: wordpress
    volumes:
      - db_data:/var/lib/mysql
volumes:
  wordpress_data:
  db_data:`,
  },
  {
    name: 'Nginx + PHP-FPM',
    description: 'Nginx web server with PHP-FPM for PHP applications',
    category: 'Web Server',
    content: `version: "3.8"
services:
  nginx:
    image: nginx:alpine
    ports:
      - "80:80"
    volumes:
      - ./html:/usr/share/nginx/html
      - ./nginx.conf:/etc/nginx/conf.d/default.conf
    depends_on:
      - php
  php:
    image: php:8.2-fpm
    volumes:
      - ./html:/var/www/html`,
  },
  {
    name: 'Node.js + Redis',
    description: 'Node.js application with Redis cache',
    category: 'Development',
    content: `version: "3.8"
services:
  app:
    build: .
    ports:
      - "3000:3000"
    environment:
      REDIS_HOST: redis
      REDIS_PORT: 6379
    depends_on:
      - redis
    volumes:
      - .:/app
      - /app/node_modules
  redis:
    image: redis:alpine
    ports:
      - "6379:6379"
    volumes:
      - redis_data:/data
volumes:
  redis_data:`,
  },
  {
    name: 'PostgreSQL',
    description: 'PostgreSQL database with persistent storage',
    category: 'Database',
    content: `version: "3.8"
services:
  db:
    image: postgres:16-alpine
    ports:
      - "5432:5432"
    environment:
      POSTGRES_USER: appuser
      POSTGRES_PASSWORD: apppassword
      POSTGRES_DB: appdb
    volumes:
      - postgres_data:/var/lib/postgresql/data
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U appuser -d appdb"]
      interval: 10s
      timeout: 5s
      retries: 5
volumes:
  postgres_data:`,
  },
  {
    name: 'LAMP Stack',
    description: 'Complete LAMP stack (Linux, Apache, MySQL, PHP)',
    category: 'Web Server',
    content: `version: "3.8"
services:
  apache:
    image: httpd:latest
    ports:
      - "80:80"
    volumes:
      - ./www:/usr/local/apache2/htdocs/
  php:
    image: php:8.2-apache
    ports:
      - "8080:80"
    volumes:
      - ./www:/var/www/html
    depends_on:
      - mysql
  mysql:
    image: mysql:8.0
    environment:
      MYSQL_ROOT_PASSWORD: rootpassword
      MYSQL_DATABASE: appdb
      MYSQL_USER: appuser
      MYSQL_PASSWORD: apppassword
    volumes:
      - mysql_data:/var/lib/mysql
volumes:
  mysql_data:`,
  },
]

export function getTemplatesByCategory(category?: string): ComposeTemplate[] {
  if (!category) return composeTemplates
  return composeTemplates.filter((t) => t.category === category)
}

export function getTemplateByName(name: string): ComposeTemplate | undefined {
  return composeTemplates.find((t) => t.name.toLowerCase() === name.toLowerCase())
}
