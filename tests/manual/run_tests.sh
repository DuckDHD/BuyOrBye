#!/bin/bash

# OpenAI Integration Test Runner
# This script runs manual tests against the real OpenAI API

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}🤖 BuyOrBye OpenAI Integration Test Runner${NC}"
echo "=" $(printf "%*s" 50 "" | tr ' ' '=')

# Check for API key
if [ -z "$OPENAI_API_KEY" ]; then
    echo -e "${RED}❌ Error: OPENAI_API_KEY environment variable not set${NC}"
    echo -e "${YELLOW}💡 Set it with: export OPENAI_API_KEY=your_api_key_here${NC}"
    echo -e "${YELLOW}💡 Or create a .env file in the project root${NC}"
    exit 1
fi

# Mask API key for display
MASKED_KEY="${OPENAI_API_KEY:0:8}$(printf "%*s" $((${#OPENAI_API_KEY} - 16)) "" | tr ' ' '*')${OPENAI_API_KEY: -8}"
echo -e "${GREEN}✅ OpenAI API Key found: ${MASKED_KEY}${NC}"

# Check if we're in the right directory
if [ ! -f "go.mod" ]; then
    echo -e "${RED}❌ Error: Please run this script from the project root directory${NC}"
    exit 1
fi

echo -e "${BLUE}📋 Available Test Scenarios:${NC}"
echo "  1. Small Purchase (\$50 gadget, low urgency)"
echo "  2. Medium Purchase (\$500 electronics, medium urgency)"
echo "  3. Large Purchase (\$2000 laptop, high urgency)"
echo "  4. Health Purchase (medical equipment, critical)"
echo "  5. Subscription (monthly service)"
echo "  6. Cache Functionality Test"
echo "  7. Response Format Validation"
echo ""

# Ask user what they want to run
echo -e "${YELLOW}Select test option:${NC}"
echo "  [a] Run all tests (recommended)"
echo "  [s] Run specific test"
echo "  [q] Quick validation test"
read -p "Choice [a/s/q]: " choice

case $choice in
    [Aa]* )
        echo -e "${GREEN}🚀 Running all OpenAI integration tests...${NC}"
        go test -v -tags=manual ./tests/manual/ -timeout=5m
        ;;
    [Ss]* )
        echo "Available test functions:"
        echo "  TestOpenAIIntegration_SmallPurchase"
        echo "  TestOpenAIIntegration_MediumPurchase"
        echo "  TestOpenAIIntegration_LargePurchase"
        echo "  TestOpenAIIntegration_HealthPurchase"
        echo "  TestOpenAIIntegration_Subscription"
        echo "  TestOpenAIIntegration_CachingFunctionality"
        echo "  TestOpenAIIntegration_ResponseFormat"
        read -p "Enter test function name: " test_func
        echo -e "${GREEN}🚀 Running specific test: ${test_func}${NC}"
        go test -v -tags=manual ./tests/manual/ -run="$test_func" -timeout=2m
        ;;
    [Qq]* )
        echo -e "${GREEN}🚀 Running quick validation test...${NC}"
        go test -v -tags=manual ./tests/manual/ -run="TestOpenAIIntegration_SmallPurchase" -timeout=1m
        ;;
    * )
        echo -e "${RED}Invalid choice. Exiting.${NC}"
        exit 1
        ;;
esac

echo ""
echo -e "${GREEN}✅ Tests completed!${NC}"

# Check if results file was created
RESULTS_FILE=$(ls openai_test_results_*.json 2>/dev/null | head -1)
if [ -n "$RESULTS_FILE" ]; then
    echo -e "${BLUE}📊 Results saved to: ${RESULTS_FILE}${NC}"
    
    # Extract key metrics from results
    if command -v jq >/dev/null 2>&1; then
        echo -e "${YELLOW}📈 Quick Summary:${NC}"
        TOTAL_TESTS=$(jq '.total_tests' "$RESULTS_FILE")
        SUCCESSFUL_TESTS=$(jq '.successful_tests' "$RESULTS_FILE")
        TOTAL_COST=$(jq '.total_cost_usd' "$RESULTS_FILE")
        
        echo "  Total Tests: $TOTAL_TESTS"
        echo "  Successful: $SUCCESSFUL_TESTS"
        echo "  Total Cost: \$$(printf '%.4f' $TOTAL_COST)"
        
        if [ "$SUCCESSFUL_TESTS" == "$TOTAL_TESTS" ]; then
            echo -e "${GREEN}  Status: All tests passed! ✅${NC}"
        else
            echo -e "${YELLOW}  Status: Some tests failed ⚠️${NC}"
        fi
    else
        echo -e "${YELLOW}💡 Install 'jq' for detailed summary analysis${NC}"
    fi
fi

echo ""
echo -e "${BLUE}🔍 Next Steps:${NC}"
echo "  1. Review the detailed results in the JSON file"
echo "  2. Check response times and costs"
echo "  3. Verify decision quality and consistency"
echo "  4. Test caching functionality"
echo ""
echo -e "${YELLOW}📚 Documentation:${NC}"
echo "  - OpenAI API costs: https://openai.com/pricing"
echo "  - GPT-4o-mini: ~\$0.15/1M input tokens, ~\$0.60/1M output tokens"
echo "  - Target cost per decision: ~\$0.004"
echo ""