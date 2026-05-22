-- +goose Up
INSERT INTO menu_items (category_id, name, description, price, available)
SELECT 1, 'Classic Burger', 'Beef patty with lettuce and tomato', 350.00, true
WHERE NOT EXISTS (
  SELECT 1 FROM menu_items WHERE category_id = 1 AND name = 'Classic Burger'
);

INSERT INTO menu_items (category_id, name, description, price, available)
SELECT 1, 'Double Burger', 'Two beef patties', 450.00, true
WHERE NOT EXISTS (
  SELECT 1 FROM menu_items WHERE category_id = 1 AND name = 'Double Burger'
);

INSERT INTO menu_items (category_id, name, description, price, available)
SELECT 2, 'Coke', '330ml can', 120.00, true
WHERE NOT EXISTS (
  SELECT 1 FROM menu_items WHERE category_id = 2 AND name = 'Coke'
);

INSERT INTO menu_items (category_id, name, description, price, available)
SELECT 2, 'Water', '500ml bottle', 80.00, true
WHERE NOT EXISTS (
  SELECT 1 FROM menu_items WHERE category_id = 2 AND name = 'Water'
);

INSERT INTO menu_items (category_id, name, description, price, available)
SELECT 3, 'Cheesecake', 'New York style', 200.00, true
WHERE NOT EXISTS (
  SELECT 1 FROM menu_items WHERE category_id = 3 AND name = 'Cheesecake'
);

-- +goose Down
DELETE FROM menu_items
WHERE (category_id, name) IN (
  (1, 'Classic Burger'),
  (1, 'Double Burger'),
  (2, 'Coke'),
  (2, 'Water'),
  (3, 'Cheesecake')
);
