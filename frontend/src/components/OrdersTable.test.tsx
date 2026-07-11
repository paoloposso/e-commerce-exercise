import { describe, test, expect } from 'vitest';
import { render, screen } from '@testing-library/react';
import { OrdersTable } from './OrdersTable';

describe('OrdersTable Component', () => {
  const createMockStore = (orders: any[] = []) => ({
    orders
  });

  test('renders empty state when there are no orders', () => {
    const store = createMockStore([]);
    render(<OrdersTable store={store} />);

    expect(screen.getByText('No orders yet')).toBeInTheDocument();
    expect(screen.getByText('Go to the catalog and purchase something!')).toBeInTheDocument();
  });

  test('renders headers and order details rows when orders exist', () => {
    const mockOrders = [
      {
        id: 'order-123456789',
        sku: 'SKU-ABC',
        customer_id: 'cust-abc-123',
        quantity: 2,
        total_price: 39.98,
        created_at: '2026-07-11T12:00:00Z'
      },
      {
        id: 'order-987654321',
        sku: 'SKU-XYZ',
        customer_id: 'cust-xyz-789',
        quantity: 1,
        total_price: 15.50,
        created_at: '' // Test fallback for missing date
      }
    ];

    const store = createMockStore(mockOrders);
    render(<OrdersTable store={store} />);

    // Check table headers
    expect(screen.getByText('Order ID')).toBeInTheDocument();
    expect(screen.getByText('Date')).toBeInTheDocument();
    expect(screen.getByText('Customer')).toBeInTheDocument();
    expect(screen.getByText('SKU')).toBeInTheDocument();
    expect(screen.getByText('Qty')).toBeInTheDocument();
    expect(screen.getByText('Total Price')).toBeInTheDocument();

    // Check order 1
    // o.id.substring(0, 8) of 'order-123456789' is 'order-12'
    expect(screen.getByText('order-12...')).toBeInTheDocument();
    expect(screen.getByText('cust-abc-123')).toBeInTheDocument();
    expect(screen.getByText('SKU-ABC')).toBeInTheDocument();
    expect(screen.getByText('2')).toBeInTheDocument();
    expect(screen.getByText('$39.98')).toBeInTheDocument();

    // Check order 2
    expect(screen.getByText('order-98...')).toBeInTheDocument();
    expect(screen.getByText('cust-xyz-789')).toBeInTheDocument();
    expect(screen.getByText('SKU-XYZ')).toBeInTheDocument();
    expect(screen.getByText('1')).toBeInTheDocument();
    expect(screen.getByText('$15.50')).toBeInTheDocument();
  });
});
