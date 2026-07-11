import { describe, test, expect, vi } from 'vitest';
import { render, screen } from '@testing-library/react';
import { Catalog } from './Catalog';

describe('Catalog Component', () => {
  const mockProducts = [
    {
      id: 'prod-1',
      sku: 'SKU-100',
      name: 'Product 1',
      description: 'Desc 1',
      category: 'Cat 1',
      price: 10.00,
      stock: 5,
      weight_kg: 1.0
    },
    {
      id: 'prod-2',
      sku: 'SKU-200',
      name: 'Product 2',
      description: 'Desc 2',
      category: 'Cat 2',
      price: 20.00,
      stock: 0,
      weight_kg: 2.0
    }
  ];

  const createMockStore = (products: any[] = []) => ({
    products,
    quantities: {},
    handleQuantityChange: vi.fn(),
    handlePurchase: vi.fn(),
    openModal: vi.fn(),
    handleDeleteProduct: vi.fn(),
  });

  test('renders empty state in shop mode when no products are present', () => {
    const store = createMockStore([]);
    render(<Catalog store={store} mode="shop" />);

    expect(screen.getByText('No products found')).toBeInTheDocument();
    expect(screen.getByText('Upload a CSV file to populate the catalog.')).toBeInTheDocument();
  });

  test('renders empty state in admin mode when no products are present', () => {
    const store = createMockStore([]);
    render(<Catalog store={store} mode="admin" />);

    expect(screen.getByText('No products found')).toBeInTheDocument();
    expect(screen.getByText('Add products manually or upload a CSV file.')).toBeInTheDocument();
  });

  test('renders grid of ProductCards when products are present', () => {
    const store = createMockStore(mockProducts);
    render(<Catalog store={store} mode="shop" />);

    expect(screen.queryByText('No products found')).not.toBeInTheDocument();
    expect(screen.getByText('Product 1')).toBeInTheDocument();
    expect(screen.getByText('Product 2')).toBeInTheDocument();
    expect(screen.getByText('SKU: SKU-100')).toBeInTheDocument();
    expect(screen.getByText('SKU: SKU-200')).toBeInTheDocument();
  });
});
