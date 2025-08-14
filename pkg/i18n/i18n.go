package i18n

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"go-clean-ddd-es-template/pkg/errors"
)

// Translator handles internationalization
type Translator struct {
	translations  map[string]map[string]string
	defaultLocale string
	mutex         sync.RWMutex
}

// NewTranslator creates a new translator
func NewTranslator(defaultLocale string) *Translator {
	return &Translator{
		translations:  make(map[string]map[string]string),
		defaultLocale: defaultLocale,
	}
}

// LoadTranslations loads translation files from a directory
func (t *Translator) LoadTranslations(translationsDir string) error {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	// Walk through the translations directory
	err := filepath.Walk(translationsDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Only process JSON files
		if !strings.HasSuffix(info.Name(), ".json") {
			return nil
		}

		// Extract locale from filename (e.g., "en.json" -> "en")
		locale := strings.TrimSuffix(info.Name(), ".json")

		// Read and parse the translation file
		data, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("failed to read translation file %s: %w", path, err)
		}

		var translations map[string]string
		if err := json.Unmarshal(data, &translations); err != nil {
			return fmt.Errorf("failed to parse translation file %s: %w", path, err)
		}

		t.translations[locale] = translations
		return nil
	})

	return err
}

// Translate translates a key to the specified locale
func (t *Translator) Translate(key string, locale string, args ...interface{}) string {
	t.mutex.RLock()
	defer t.mutex.RUnlock()

	// Try to get translation for the specified locale
	translation, exists := t.getTranslation(key, locale)
	if !exists {
		// Fallback to default locale
		translation, exists = t.getTranslation(key, t.defaultLocale)
		if !exists {
			// Return the key if no translation found
			return key
		}
	}

	// Format the translation with arguments if provided
	if len(args) > 0 {
		return fmt.Sprintf(translation, args...)
	}

	return translation
}

// getTranslation gets a translation for a specific locale
func (t *Translator) getTranslation(key string, locale string) (string, bool) {
	localeTranslations, exists := t.translations[locale]
	if !exists {
		return "", false
	}

	translation, exists := localeTranslations[key]
	return translation, exists
}

// GetSupportedLocales returns a list of supported locales
func (t *Translator) GetSupportedLocales() []string {
	t.mutex.RLock()
	defer t.mutex.RUnlock()

	locales := make([]string, 0, len(t.translations))
	for locale := range t.translations {
		locales = append(locales, locale)
	}
	return locales
}

// IsLocaleSupported checks if a locale is supported
func (t *Translator) IsLocaleSupported(locale string) bool {
	t.mutex.RLock()
	defer t.mutex.RUnlock()

	_, exists := t.translations[locale]
	return exists
}

// TranslateError translates an AppError to the specified locale
func (t *Translator) TranslateError(err *errors.AppError, locale string) *errors.AppError {
	if err == nil {
		return nil
	}

	// Try to translate the error message
	translatedMessage := t.Translate(string(err.Code), locale)
	if translatedMessage != string(err.Code) {
		// If translation found, update the error message
		err.Message = translatedMessage
	}

	// Set the locale
	err.Locale = locale

	return err
}

// Global translator instance
var (
	globalTranslator     *Translator
	globalTranslatorOnce sync.Once
)

// GetGlobalTranslator returns the global translator instance
func GetGlobalTranslator() *Translator {
	globalTranslatorOnce.Do(func() {
		globalTranslator = NewTranslator("en")
	})
	return globalTranslator
}

// SetGlobalTranslator sets the global translator instance
func SetGlobalTranslator(translator *Translator) {
	globalTranslator = translator
}

// T is a shorthand for translating using the global translator
func T(key string, locale string, args ...interface{}) string {
	return GetGlobalTranslator().Translate(key, locale, args...)
}

// TE is a shorthand for translating errors using the global translator
func TE(err *errors.AppError, locale string) *errors.AppError {
	return GetGlobalTranslator().TranslateError(err, locale)
}
