// Swap Manager Class
class SwapManager {
    constructor() {
        this.init();
    }

    init() {
        this.setupEventListeners();
        this.setupNavbarScroll();
        this.loadCurrencies();
    }

    setupEventListeners() {
        // Amount input listener
        const fromAmount = document.getElementById('fromAmount');
        if (fromAmount) {
            fromAmount.addEventListener('input', this.handleAmountChange.bind(this));
        }

        // Currency select listeners
        const fromSelect = document.getElementById('fromSelect');
        const toSelect = document.getElementById('toSelect');
        
        if (fromSelect) {
            fromSelect.addEventListener('change', this.updateFromIcon.bind(this));
        }
        
        if (toSelect) {
            toSelect.addEventListener('change', this.updateToIcon.bind(this));
        }

        // Smooth scroll for anchor links
        document.querySelectorAll('a[href^="#"]').forEach(anchor => {
            anchor.addEventListener('click', function (e) {
                e.preventDefault();
                const target = document.querySelector(this.getAttribute('href'));
                if (target) {
                    target.scrollIntoView({
                        behavior: 'smooth',
                        block: 'start'
                    });
                }
            });
        });
    }

    setupNavbarScroll() {
        window.addEventListener('scroll', function() {
            const navbar = document.getElementById('navbar');
            if (navbar) {
                if (window.scrollY > 50) {
                    navbar.classList.add('scrolled');
                } else {
                    navbar.classList.remove('scrolled');
                }
            }
        });
    }

    loadCurrencies() {
        console.log('Loading currencies...');
        
        // Check if we're using templates or static HTML
        const fromSelect = document.getElementById('fromSelect');
        const toSelect = document.getElementById('toSelect');
        
        if (!fromSelect || !toSelect) {
            console.log('Select elements not found');
            return;
        }
        
        // If using templates, currencies are loaded via HTMX
        // If using static HTML, load via API
        if (window.location.pathname === '/index.html' || !window.htmx) {
            // Static HTML fallback - use legacy API
            fetch('/api/currencies', {
                headers: {
                    'HX-Request': 'true'
                }
            })
            .then(response => response.text())
            .then(html => {
                fromSelect.innerHTML = html;
                toSelect.innerHTML = html;
                
                fromSelect.value = 'btc';
                toSelect.value = 'eth';
                this.updateFromIcon();
                this.updateToIcon();
            })
            .catch(error => {
                console.error('Error loading currencies:', error);
                this.loadFallbackCurrencies();
            });
        }
    }
    
    loadFallbackCurrencies() {
        const fallback = `
            <option value="btc">BTC - Bitcoin</option>
            <option value="eth">ETH - Ethereum</option>
            <option value="usdt">USDT - Tether</option>
            <option value="usdc">USDC - USD Coin</option>
            <option value="bnb">BNB - Binance Coin</option>
        `;
        
        const fromSelect = document.getElementById('fromSelect');
        const toSelect = document.getElementById('toSelect');
        
        if (fromSelect && toSelect) {
            fromSelect.innerHTML = fallback;
            toSelect.innerHTML = fallback;
            fromSelect.value = 'btc';
            toSelect.value = 'eth';
            this.updateFromIcon();
            this.updateToIcon();
        }
    }

    handleAmountChange(event) {
        const amount = parseFloat(event.target.value);
        const button = document.getElementById('swapButton');
        
        if (button) {
            if (amount && amount > 0) {
                button.disabled = false;
                const span = button.querySelector('span');
                if (span) {
                    span.textContent = 'Get Best Quote';
                }
            } else {
                button.disabled = true;
                const span = button.querySelector('span');
                if (span) {
                    span.textContent = 'Enter Amount To Preview Swap';
                }
                const toAmount = document.getElementById('toAmount');
                if (toAmount) {
                    toAmount.value = '';
                }
            }
        }
    }

    updateFromIcon() {
        const fromSelect = document.getElementById('fromSelect');
        const fromIcon = document.getElementById('fromTokenIcon');
        if (fromSelect && fromIcon) {
            fromIcon.textContent = fromSelect.value.toUpperCase().slice(0, 3);
        }
    }

    updateToIcon() {
        const toSelect = document.getElementById('toSelect');
        const toIcon = document.getElementById('toTokenIcon');
        if (toSelect && toIcon) {
            toIcon.textContent = toSelect.value.toUpperCase().slice(0, 3);
        }
    }
}

// Global functions that can be called from HTML
function swapPairs() {
    const fromSelect = document.getElementById('fromSelect');
    const toSelect = document.getElementById('toSelect');
    const fromAmount = document.getElementById('fromAmount');
    const toAmount = document.getElementById('toAmount');
    
    if (fromSelect && toSelect) {
        const tempValue = fromSelect.value;
        fromSelect.value = toSelect.value;
        toSelect.value = tempValue;
        
        const fromIcon = document.getElementById('fromTokenIcon');
        const toIcon = document.getElementById('toTokenIcon');
        
        if (fromIcon) {
            fromIcon.textContent = fromSelect.value.toUpperCase().slice(0, 3);
        }
        if (toIcon) {
            toIcon.textContent = toSelect.value.toUpperCase().slice(0, 3);
        }
        
        if (toAmount) {
            toAmount.value = '';
        }
        
        if (fromAmount && fromAmount.value && window.htmx) {
            htmx.trigger(fromAmount, 'keyup');
        }
    }
}

function executeSwap() {
    const from = document.getElementById('fromSelect').value;
    const to = document.getElementById('toSelect').value;
    const amount = document.getElementById('fromAmount').value;
    const address = document.getElementById('toAddress').value;
    const exchange = document.getElementById('swapButton').getAttribute('data-exchange') || '';
    
    if (!address) {
        alert('Please enter your receiving wallet address');
        document.querySelector('details').open = true;
        document.getElementById('toAddress').focus();
        return;
    }
    
    const formData = new FormData();
    formData.append('from', from);
    formData.append('to', to);
    formData.append('amount', amount);
    formData.append('to_address', address);
    formData.append('exchange', exchange);
    
    fetch('/htmx/swap', {
        method: 'POST',
        body: formData,
        headers: {
            'HX-Request': 'true'
        }
    })
    .then(response => response.text())
    .then(html => {
        const swapCard = document.querySelector('.swap-card');
        if (swapCard) {
            swapCard.innerHTML = html;
        }
    })
    .catch(error => {
        alert('Error: ' + error);
    });
}

function toggleMobileMenu() {
    const menu = document.getElementById('mobileMenu');
    if (menu) {
        menu.classList.toggle('active');
    }
}

function selectExchange(exchange) {
    const button = document.getElementById('swapButton');
    if (button) {
        const buttonText = button.querySelector('span') || button;
        buttonText.textContent = 'Swap via ' + exchange;
        button.setAttribute('data-exchange', exchange);
    }
}

// Initialize when DOM is ready
document.addEventListener('DOMContentLoaded', () => {
    new SwapManager();
});