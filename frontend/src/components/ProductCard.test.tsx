import { describe, test, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import { ProductCard } from './ProductCard';

describe('ProductCard Component', () => {
  const mockProduct = {
    id: 'prod-1',
    sku: 'SKU-100',
    name: 'Sample Product',
    description: 'This is a sample product description',
    category: 'Electronics',
    price: 19.99,
    stock: 5,
    weight_kg: 1.2
  };

  const createMockStore = (overrides = {}) => ({
    quantities: {},
    handleQuantityChange: vi.fn(),
    handlePurchase: vi.fn(),
    openModal: vi.fn(),
    handleDeleteProduct: vi.fn(),
    ...overrides,
  });

  test('renders product basic information', () => {
    const store = createMockStore();
    render(<ProductCard p={mockProduct} index={0} store={store} mode="shop" />);

    expect(screen.getByText('Sample Product')).toBeInTheDocument();
    expect(screen.getByText('SKU: SKU-100')).toBeInTheDocument();
    expect(screen.getByText('This is a sample product description')).toBeInTheDocument();
    expect(screen.getByText('Electronics')).toBeInTheDocument();
    expect(screen.getByText('$19.99')).toBeInTheDocument();
    expect(screen.getByText('5 in stock')).toBeInTheDocument();
  });

  test('shows default category "General" if not provided', () => {
    const store = createMockStore();
    const noCatProduct = { ...mockProduct, category: '' };
    render(<ProductCard p={noCatProduct} index={0} store={store} mode="shop" />);

    expect(screen.getByText('General')).toBeInTheDocument();
  });

  test('shows out of stock text and disables action buttons when stock is 0', () => {
    const store = createMockStore();
    const outOfStockProduct = { ...mockProduct, stock: 0 };
    render(<ProductCard p={outOfStockProduct} index={0} store={store} mode="shop" />);

    expect(screen.getByText('Out of stock')).toBeInTheDocument();
    
    const buyButton = screen.getByRole('button', { name: /buy now/i });
    expect(buyButton).toBeDisabled();

    const [minusBtn, plusBtn] = screen.getAllByRole('button', { name: /[+-]/ });
    expect(minusBtn).toBeDisabled();
    expect(plusBtn).toBeDisabled();
    expect(screen.getByText('0')).toBeInTheDocument();
  });

  test('handles quantity selector clicks in shop mode', () => {
    const store = createMockStore({
      quantities: { 'SKU-100': 2 }
    });
    render(<ProductCard p={mockProduct} index={0} store={store} mode="shop" />);

    // Total price is price * qty = 19.99 * 2 = 39.98
    expect(screen.getByText('$39.98')).toBeInTheDocument();
    expect(screen.getByText('2')).toBeInTheDocument();

    const minusBtn = screen.getByRole('button', { name: '-' });
    const plusBtn = screen.getByRole('button', { name: '+' });

    fireEvent.click(minusBtn);
    expect(store.handleQuantityChange).toHaveBeenCalledWith('SKU-100', -1, 5);

    fireEvent.click(plusBtn);
    expect(store.handleQuantityChange).toHaveBeenCalledWith('SKU-100', 1, 5);
  });

  test('disables minus button when quantity is 1', () => {
    const store = createMockStore({
      quantities: { 'SKU-100': 1 }
    });
    render(<ProductCard p={mockProduct} index={0} store={store} mode="shop" />);

    const minusBtn = screen.getByRole('button', { name: '-' });
    expect(minusBtn).toBeDisabled();
  });

  test('disables plus button when quantity reaches stock limit', () => {
    const store = createMockStore({
      quantities: { 'SKU-100': 5 }
    });
    render(<ProductCard p={mockProduct} index={0} store={store} mode="shop" />);

    const plusBtn = screen.getByRole('button', { name: '+' });
    expect(plusBtn).toBeDisabled();
  });

  test('calls handlePurchase when Buy Now is clicked', () => {
    const store = createMockStore();
    render(<ProductCard p={mockProduct} index={0} store={store} mode="shop" />);

    const buyButton = screen.getByRole('button', { name: /buy now/i });
    fireEvent.click(buyButton);

    expect(store.handlePurchase).toHaveBeenCalledWith(mockProduct);
  });

  test('renders action buttons in admin mode and triggers edit/delete', () => {
    const store = createMockStore();
    render(<ProductCard p={mockProduct} index={0} store={store} mode="admin" />);

    // Should render edit (✏️) and delete (🗑️) buttons
    const editBtn = screen.getByTitle('Edit Product');
    const deleteBtn = screen.getByTitle('Delete Product');

    expect(editBtn).toBeInTheDocument();
    expect(deleteBtn).toBeInTheDocument();

    fireEvent.click(editBtn);
    expect(store.openModal).toHaveBeenCalledWith(mockProduct);

    fireEvent.click(deleteBtn);
    expect(store.handleDeleteProduct).toHaveBeenCalledWith(mockProduct);

    // Quantity selector should not be present in admin mode
    expect(screen.queryByRole('button', { name: /buy now/i })).not.toBeInTheDocument();
  });
});
