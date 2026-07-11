import { describe, test, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import { ProductModal } from './ProductModal';

describe('ProductModal Component', () => {
  const createMockStore = (overrides = {}) => ({
    isModalOpen: true,
    setIsModalOpen: vi.fn(),
    editingProduct: null,
    handleSaveProduct: vi.fn(),
    formName: '',
    setFormName: vi.fn(),
    formSku: '',
    setFormSku: vi.fn(),
    formCategory: '',
    setFormCategory: vi.fn(),
    formDescription: '',
    setFormDescription: vi.fn(),
    formPrice: 0.00,
    handlePriceChange: vi.fn(),
    formStock: 0,
    setFormStock: vi.fn(),
    formWeight: 0.00,
    setFormWeight: vi.fn(),
    fieldErrors: {},
    ...overrides,
  });

  test('does not render if isModalOpen is false', () => {
    const store = createMockStore({ isModalOpen: false });
    const { container } = render(<ProductModal store={store} />);
    expect(container.firstChild).toBeNull();
  });

  test('renders "Add New Product" heading and empty fields in creation mode', () => {
    const store = createMockStore();
    render(<ProductModal store={store} />);

    expect(screen.getByRole('heading', { name: 'Add New Product' })).toBeInTheDocument();
    
    expect(screen.getByLabelText('Product Name *')).toHaveValue('');
    expect(screen.getByLabelText('SKU *')).toHaveValue('');
    expect(screen.getByLabelText('Category')).toHaveValue('');
    expect(screen.getByLabelText('Description')).toHaveValue('');
    expect(screen.getByLabelText('Price ($) *')).toHaveValue('0.00');
    expect(screen.getByLabelText('Stock *')).toHaveValue(0);
    expect(screen.getByLabelText('Weight (kg)')).toHaveValue(0);
  });

  test('renders "Edit Product" heading and pre-filled fields in editing mode', () => {
    const store = createMockStore({
      editingProduct: { id: 'prod-1', sku: 'SKU-ABC' },
      formName: 'Awesome Phone',
      formSku: 'AWESOME-PHONE',
      formCategory: 'Smartphones',
      formDescription: 'Best phone ever',
      formPrice: 999.99,
      formStock: 15,
      formWeight: 0.22,
    });
    render(<ProductModal store={store} />);

    expect(screen.getByRole('heading', { name: 'Edit Product' })).toBeInTheDocument();
    
    expect(screen.getByLabelText('Product Name *')).toHaveValue('Awesome Phone');
    expect(screen.getByLabelText('SKU *')).toHaveValue('AWESOME-PHONE');
    expect(screen.getByLabelText('Category')).toHaveValue('Smartphones');
    expect(screen.getByLabelText('Description')).toHaveValue('Best phone ever');
    expect(screen.getByLabelText('Price ($) *')).toHaveValue('999.99');
    expect(screen.getByLabelText('Stock *')).toHaveValue(15);
    expect(screen.getByLabelText('Weight (kg)')).toHaveValue(0.22);
  });

  test('calls store handlers when inputs are changed', () => {
    const store = createMockStore();
    render(<ProductModal store={store} />);

    // Name input change
    fireEvent.change(screen.getByLabelText('Product Name *'), { target: { value: 'New Name' } });
    expect(store.setFormName).toHaveBeenCalledWith('New Name');

    // SKU input change
    fireEvent.change(screen.getByLabelText('SKU *'), { target: { value: 'NEW-SKU' } });
    expect(store.setFormSku).toHaveBeenCalledWith('NEW-SKU');

    // Category input change
    fireEvent.change(screen.getByLabelText('Category'), { target: { value: 'New Category' } });
    expect(store.setFormCategory).toHaveBeenCalledWith('New Category');

    // Description input change
    fireEvent.change(screen.getByLabelText('Description'), { target: { value: 'New description details' } });
    expect(store.setFormDescription).toHaveBeenCalledWith('New description details');

    // Price input change (uses handlePriceChange)
    fireEvent.change(screen.getByLabelText('Price ($) *'), { target: { value: '12.50' } });
    expect(store.handlePriceChange).toHaveBeenCalledWith('12.50');

    // Stock input change (uses setFormStock with number conversion)
    fireEvent.change(screen.getByLabelText('Stock *'), { target: { value: '10' } });
    expect(store.setFormStock).toHaveBeenCalledWith(10);

    // Weight input change (uses setFormWeight with number conversion)
    fireEvent.change(screen.getByLabelText('Weight (kg)'), { target: { value: '2.5' } });
    expect(store.setFormWeight).toHaveBeenCalledWith(2.5);
  });

  test('displays validation errors when they are present in the store', () => {
    const store = createMockStore({
      fieldErrors: {
        name: 'Product Name is required',
        sku: 'SKU can only contain letters, numbers, hyphens, and underscores',
        price: 'Price cannot be negative',
      }
    });
    render(<ProductModal store={store} />);

    expect(screen.getByText('Product Name is required')).toBeInTheDocument();
    expect(screen.getByText('SKU can only contain letters, numbers, hyphens, and underscores')).toBeInTheDocument();
    expect(screen.getByText('Price cannot be negative')).toBeInTheDocument();
  });

  test('triggers handleSaveProduct when form is submitted', () => {
    const store = createMockStore();
    render(<ProductModal store={store} />);

    const form = screen.getByRole('button', { name: 'Save Product' }).closest('form');
    expect(form).toBeInTheDocument();

    fireEvent.submit(form!);
    expect(store.handleSaveProduct).toHaveBeenCalled();
  });

  test('calls setIsModalOpen(false) when close button, cancel button, or backdrop is clicked', () => {
    const store = createMockStore();
    const { container } = render(<ProductModal store={store} />);

    // Click close button (×)
    fireEvent.click(screen.getByText('×'));
    expect(store.setIsModalOpen).toHaveBeenCalledWith(false);

    // Click cancel button
    fireEvent.click(screen.getByRole('button', { name: 'Cancel' }));
    expect(store.setIsModalOpen).toHaveBeenCalledWith(false);

    // Click backdrop
    const backdrop = container.querySelector('.modal-backdrop');
    expect(backdrop).toBeInTheDocument();
    fireEvent.click(backdrop!);
    expect(store.setIsModalOpen).toHaveBeenCalledWith(false);
  });
});
