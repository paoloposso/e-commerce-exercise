import { useState, useEffect, useRef } from 'react';
import { api } from '../services/api';
import type { Product, PurchaseResult } from '../services/api';

function generateUUID() {
  return crypto.randomUUID();
}

export function useAppStore() {
  const [activeTab, setActiveTab] = useState<'shop' | 'inventory' | 'orders'>('shop');
  
  const [products, setProducts] = useState<Product[]>([]);
  const [orders, setOrders] = useState<PurchaseResult[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  
  const [search, setSearch] = useState('');
  const [isImporting, setIsImporting] = useState(false);
  const fileInputRef = useRef<HTMLInputElement>(null);

  const [toast, setToast] = useState<{message: string, type: 'success' | 'error'} | null>(null);

  const [isModalOpen, setIsModalOpen] = useState(false);
  const [editingProduct, setEditingProduct] = useState<Product | null>(null);
  const [formName, setFormName] = useState('');
  const [formSku, setFormSku] = useState('');
  const [formDescription, setFormDescription] = useState('');
  const [formCategory, setFormCategory] = useState('');
  const [formPrice, setFormPrice] = useState(0);
  const [formStock, setFormStock] = useState(0);
  const [formWeight, setFormWeight] = useState(0);

  const [quantities, setQuantities] = useState<Record<string, number>>({});
  const [fieldErrors, setFieldErrors] = useState<Record<string, string>>({});

  const [isInitialLoad, setIsInitialLoad] = useState(true);

  const fetchData = async () => {
    try {
      if (isInitialLoad) {
        setLoading(true);
      }
      setError(null);
      if (activeTab === 'shop' || activeTab === 'inventory') {
        const data = await api.listProducts(search);
        setProducts(data || []);
      } else {
        const data = await api.listOrders();
        setOrders(data || []);
      }
    } catch (err: any) {
      setError(err.message);
    } finally {
      setLoading(false);
      setIsInitialLoad(false);
    }
  };

  useEffect(() => {
    const delayDebounceFn = setTimeout(() => {
      fetchData();
    }, 300);
    return () => clearTimeout(delayDebounceFn);
  }, [search, activeTab]);

  const showToast = (message: string, type: 'success' | 'error') => {
    setToast({ message, type });
    setTimeout(() => setToast(null), 5000);
  };

  const handleQuantityChange = (sku: string, delta: number, max: number) => {
    setQuantities(prev => {
      const current = prev[sku] || 1;
      const next = Math.max(1, Math.min(max, current + delta));
      return { ...prev, [sku]: next };
    });
  };

  const handlePurchase = async (product: Product) => {
    try {
      const idempotencyKey = generateUUID();
      const customerId = 'cust-demo-123';
      const quantityToBuy = quantities[product.sku] || 1;
      
      await api.purchase(product.sku, customerId, quantityToBuy, product.price, idempotencyKey);
      
      showToast(`Successfully purchased ${quantityToBuy}x ${product.name}!`, 'success');
      setQuantities(prev => ({ ...prev, [product.sku]: 1 }));
      fetchData();
    } catch (err: any) {
      showToast(err.message, 'error');
    }
  };

  const handleFileUpload = async (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (!file) return;

    try {
      setIsImporting(true);
      const report = await api.importCsv(file);
      showToast(`Imported ${report.imported_rows} new products and updated ${report.updated_rows}.`, 'success');
      fetchData();
    } catch (err: any) {
      showToast(err.message, 'error');
    } finally {
      setIsImporting(false);
      if (fileInputRef.current) {
        fileInputRef.current.value = '';
      }
    }
  };

  const openModal = (product: Product | null = null) => {
    setEditingProduct(product);
    if (product) {
      setFormName(product.name);
      setFormSku(product.sku);
      setFormDescription(product.description || '');
      setFormCategory(product.category || '');
      setFormPrice(product.price);
      setFormStock(product.stock);
      setFormWeight(product.weight_kg);
    } else {
      setFormName('');
      setFormSku('');
      setFormDescription('');
      setFormCategory('');
      setFormPrice(0);
      setFormStock(0);
      setFormWeight(0);
    }
    setFieldErrors({});
    setIsModalOpen(true);
  };

  const handleSaveProduct = async (e: React.FormEvent) => {
    e.preventDefault();

    const trimmedName = formName.trim();
    const trimmedSku = formSku.trim();
    const errors: Record<string, string> = {};

    if (!trimmedName) {
      errors.name = 'Product Name is required';
    }
    if (!trimmedSku) {
      errors.sku = 'SKU is required';
    } else {
      const skuRegex = /^[a-zA-Z0-9-_]+$/;
      if (!skuRegex.test(trimmedSku)) {
        errors.sku = 'SKU can only contain letters, numbers, hyphens, and underscores';
      }
    }
    if (formPrice < 0) {
      errors.price = 'Price cannot be negative';
    }
    if (formStock < 0) {
      errors.stock = 'Stock cannot be negative';
    }
    if (formWeight < 0) {
      errors.weight = 'Weight cannot be negative';
    }

    if (Object.keys(errors).length > 0) {
      setFieldErrors(errors);
      showToast('Please correct the validation errors', 'error');
      return;
    }
    setFieldErrors({});

    try {
      const payload = {
        name: trimmedName,
        sku: trimmedSku,
        description: formDescription.trim(),
        category: formCategory.trim(),
        price: Number(formPrice),
        stock: Number(formStock),
        weight_kg: Number(formWeight),
      };

      if (editingProduct) {
        await api.updateProduct(editingProduct.id, payload);
        showToast('Product updated successfully!', 'success');
      } else {
        await api.createProduct(payload);
        showToast('Product created successfully!', 'success');
      }
      setIsModalOpen(false);
      fetchData();
    } catch (err: any) {
      showToast(err.message, 'error');
    }
  };

  const handleDeleteProduct = async (product: Product) => {
    if (!window.confirm(`Are you sure you want to delete ${product.name}?`)) return;
    try {
      await api.deleteProduct(product.id);
      showToast('Product deleted successfully!', 'success');
      fetchData();
    } catch (err: any) {
      showToast(err.message, 'error');
    }
  };

  const handlePriceChange = (val: string) => {
    const digits = val.replace(/\D/g, '');
    if (!digits) {
      setFormPrice(0);
      return;
    }
    const numericValue = parseInt(digits, 10) / 100;
    setFormPrice(numericValue);
  };

  return {
    activeTab, setActiveTab,
    products, orders,
    loading, error,
    search, setSearch,
    isImporting, fileInputRef,
    toast,
    isModalOpen, setIsModalOpen,
    editingProduct,
    formName, setFormName,
    formSku, setFormSku,
    formDescription, setFormDescription,
    formCategory, setFormCategory,
    formPrice, setFormPrice,
    formStock, setFormStock,
    formWeight, setFormWeight,
    quantities,
    handleQuantityChange,
    handlePurchase,
    handleFileUpload,
    openModal,
    handleSaveProduct,
    handleDeleteProduct,
    handlePriceChange,
    fieldErrors
  };
}
