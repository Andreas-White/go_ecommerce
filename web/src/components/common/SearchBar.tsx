import React, { useState, useEffect, useRef, useCallback } from 'react';
import { useRouter } from 'next/navigation';
import Image from 'next/image';
import Input from '../ui/Input';
import './SearchBar.css';
import { Button } from '@/components/ui';
import { api } from '@/lib/api';

interface ProductSuggestion {
  id: string;
  name: string;
  image_url?: string;
  price?: number;
  description?: string;
  stock?: number;
  category?: string;
}

export default function SearchBar({
  value,
  onChange,
  onSubmit,
}: {
  value: string;
  onChange: (v: string) => void;
  onSubmit?: (value: string) => void;
}) {
  const router = useRouter();
  const [suggestions, setSuggestions] = useState<ProductSuggestion[]>([]);
  const [showDropdown, setShowDropdown] = useState(false);
  const [isLoading, setIsLoading] = useState(false);
  const [highlightedIndex, setHighlightedIndex] = useState(-1);
  const dropdownRef = useRef<HTMLDivElement>(null);
  const debounceRef = useRef<NodeJS.Timeout | null>(null);
  const fetchIdRef = useRef<number>(0);

  const safeSuggestions = suggestions || [];

  const fetchSuggestions = useCallback(async (searchTerm: string) => {
    const currentFetchId = ++fetchIdRef.current;

    setIsLoading(true);
    try {
      const data = await api.get<ProductSuggestion[]>(
        `/products?search=${encodeURIComponent(searchTerm)}&limit=5`
      );
      if (currentFetchId === fetchIdRef.current) {
        setSuggestions(data || []);
        setShowDropdown(true);
        setHighlightedIndex(-1);
      }
    } catch (error) {
      if (currentFetchId === fetchIdRef.current) {
        console.error('Failed to fetch suggestions:', error);
        setSuggestions([]);
      }
    } finally {
      if (currentFetchId === fetchIdRef.current) {
        setIsLoading(false);
      }
    }
  }, []);

  useEffect(() => {
    if (debounceRef.current) {
      clearTimeout(debounceRef.current);
    }

    if (value.trim().length > 0) {
      debounceRef.current = setTimeout(() => {
        fetchSuggestions(value.trim());
      }, 300);
    } else {
      setSuggestions([]);
      setShowDropdown(false);
      setHighlightedIndex(-1);
    }

    return () => {
      if (debounceRef.current) {
        clearTimeout(debounceRef.current);
      }
    };
  }, [value, fetchSuggestions]);

  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      if (
        dropdownRef.current &&
        !dropdownRef.current.contains(event.target as Node)
      ) {
        setShowDropdown(false);
      }
    };

    document.addEventListener('mousedown', handleClickOutside);
    return () => {
      document.removeEventListener('mousedown', handleClickOutside);
    };
  }, []);

  const handleSuggestionClick = (id: string) => {
    setShowDropdown(false);
    setSuggestions([]);
    onChange('');
    router.push(`/product/${id}`);
  };

  const handleKeyDown = (e: React.KeyboardEvent<HTMLInputElement>) => {
    if (!showDropdown || safeSuggestions.length === 0) {
      if (e.key === 'Enter') {
        onSubmit?.(value);
      }
      return;
    }

    switch (e.key) {
      case 'ArrowDown':
        e.preventDefault();
        setHighlightedIndex((prev) =>
          prev < safeSuggestions.length - 1 ? prev + 1 : prev
        );
        break;
      case 'ArrowUp':
        e.preventDefault();
        setHighlightedIndex((prev) => (prev > 0 ? prev - 1 : -1));
        break;
      case 'Enter':
        e.preventDefault();
        if (
          highlightedIndex >= 0 &&
          highlightedIndex < safeSuggestions.length
        ) {
          handleSuggestionClick(safeSuggestions[highlightedIndex].id);
        } else {
          onSubmit?.(value);
        }
        break;
      case 'Escape':
        setShowDropdown(false);
        setHighlightedIndex(-1);
        break;
    }
  };

  return (
    <form
      onSubmit={(e) => {
        e.preventDefault();
        setShowDropdown(false);
        onSubmit?.(value);
      }}
      className="search-bar-form"
    >
      <div className="search-bar-input-wrapper">
        <Input
          type="text"
          value={value}
          onChange={(e) => onChange(e.target.value)}
          placeholder="Search products..."
          className="search-bar-input"
          onKeyDown={handleKeyDown}
          onFocus={() => {
            if (safeSuggestions.length > 0) {
              setShowDropdown(true);
            }
          }}
        />
        {value.length > 0 && (
          <button
            type="button"
            onClick={() => {
              onChange('');
              setSuggestions([]);
              setShowDropdown(false);
            }}
            className="search-bar-clear-button visible"
            aria-label="Clear search"
          >
            ✕
          </button>
        )}
        {showDropdown && (
          <div className="search-bar-dropdown" ref={dropdownRef}>
            {isLoading && <div className="search-bar-loading">Loading...</div>}
            {!isLoading &&
              safeSuggestions.length === 0 &&
              value.trim().length > 0 && (
                <div className="search-bar-no-results">No products found</div>
              )}
            {!isLoading &&
              safeSuggestions.map((suggestion, index) => (
                <div
                  key={suggestion.id}
                  className={`search-bar-suggestion ${
                    index === highlightedIndex ? 'highlighted' : ''
                  }`}
                  onClick={() => handleSuggestionClick(suggestion.id)}
                  onMouseEnter={() => setHighlightedIndex(index)}
                >
                  {suggestion.image_url && (
                    <div className="search-bar-suggestion-image-wrapper">
                      <Image
                        src={suggestion.image_url}
                        alt=""
                        fill
                        sizes="40px"
                        className="search-bar-suggestion-image"
                      />
                    </div>
                  )}
                  <span className="search-bar-suggestion-name">
                    {suggestion.name}
                  </span>
                </div>
              ))}
          </div>
        )}
      </div>
      <Button type="submit" className="search-bar-button" aria-label="Search">
        🔍
      </Button>
    </form>
  );
}
