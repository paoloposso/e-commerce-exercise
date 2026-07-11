import { describe, test, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import { Header } from './Header';

describe('Header Component', () => {
  const createMockStore = (overrides = {}) => ({
    activeTab: 'shop',
    setActiveTab: vi.fn(),
    search: '',
    setSearch: vi.fn(),
    openModal: vi.fn(),
    fileInputRef: { current: { click: vi.fn() } },
    isImporting: false,
    handleFileUpload: vi.fn(),
    ...overrides,
  });

  test('renders logo and navigation tabs', () => {
    const store = createMockStore();
    render(<Header store={store} />);

    expect(screen.getByText('Store')).toBeInTheDocument();
    expect(screen.getByText('Shop')).toBeInTheDocument();
    expect(screen.getByText('Inventory')).toBeInTheDocument();
    expect(screen.getByText('Order History')).toBeInTheDocument();
  });

  test('triggers setActiveTab when tabs are clicked', () => {
    const store = createMockStore();
    render(<Header store={store} />);

    fireEvent.click(screen.getByText('Inventory'));
    expect(store.setActiveTab).toHaveBeenCalledWith('inventory');

    fireEvent.click(screen.getByText('Order History'));
    expect(store.setActiveTab).toHaveBeenCalledWith('orders');
  });

  test('shows search input in shop view and triggers setSearch', () => {
    const store = createMockStore({ activeTab: 'shop', search: 'iphone' });
    render(<Header store={store} />);

    const input = screen.getByPlaceholderText('Search catalog...') as HTMLInputElement;
    expect(input).toBeInTheDocument();
    expect(input.value).toBe('iphone');

    fireEvent.change(input, { target: { value: 'ipad' } });
    expect(store.setSearch).toHaveBeenCalledWith('ipad');
  });

  test('shows search input, Add Product, and Upload CSV in inventory view', () => {
    const store = createMockStore({ activeTab: 'inventory' });
    render(<Header store={store} />);

    expect(screen.getByPlaceholderText('Search inventory...')).toBeInTheDocument();
    expect(screen.getByText('+ Add Product')).toBeInTheDocument();
    expect(screen.getByText('Upload CSV')).toBeInTheDocument();
  });

  test('triggers openModal when + Add Product is clicked in inventory view', () => {
    const store = createMockStore({ activeTab: 'inventory' });
    render(<Header store={store} />);

    fireEvent.click(screen.getByText('+ Add Product'));
    expect(store.openModal).toHaveBeenCalled();
  });

  test('triggers file upload click when Upload CSV is clicked', () => {
    const store = createMockStore({ activeTab: 'inventory' });
    const clickSpy = vi.spyOn(HTMLInputElement.prototype, 'click').mockImplementation(() => {});
    
    render(<Header store={store} />);

    fireEvent.click(screen.getByText('Upload CSV'));
    expect(clickSpy).toHaveBeenCalled();
    
    clickSpy.mockRestore();
  });

  test('shows importing state when isImporting is true', () => {
    const store = createMockStore({ activeTab: 'inventory', isImporting: true });
    render(<Header store={store} />);

    const uploadBtn = screen.getByText('Importing...');
    expect(uploadBtn).toBeInTheDocument();
    expect(uploadBtn).toBeDisabled();
  });
});
