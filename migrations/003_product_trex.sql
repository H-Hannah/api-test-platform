-- 将产品分组统一为 Trex，移除示例项目 B
UPDATE products SET name = 'Trex' WHERE id = 1;
DELETE FROM products WHERE id = 2;
