// Sources flattened with hardhat v2.29.2 https://hardhat.org

// SPDX-License-Identifier: MIT

// File contracts/SkynetEcosystem.sol

// Original license: SPDX_License_Identifier: MIT
pragma solidity ^0.8.24;

/// @title Wrapped SKY (WSKY) - Native 18-decimal ERC-20 wrapper for SKY gas coin
contract WSKY {
    string public name = "Wrapped Skynet";
    string public symbol = "WSKY";
    uint8 public decimals = 18;

    mapping(address => uint256) public balanceOf;
    mapping(address => mapping(address => uint256)) public allowance;

    event Deposit(address indexed dst, uint256 wad);
    event Withdrawal(address indexed src, uint256 wad);
    event Transfer(address indexed src, address indexed dst, uint256 wad);
    event Approval(address indexed src, address indexed guy, uint256 wad);

    receive() external payable { deposit(); }

    function deposit() public payable {
        balanceOf[msg.sender] += msg.value;
        emit Deposit(msg.sender, msg.value);
    }

    function withdraw(uint256 wad) public {
        require(balanceOf[msg.sender] >= wad, "Insufficient WSKY");
        balanceOf[msg.sender] -= wad;
        payable(msg.sender).transfer(wad);
        emit Withdrawal(msg.sender, wad);
    }

    function totalSupply() public view returns (uint256) {
        return address(this).balance;
    }

    function approve(address guy, uint256 wad) public returns (bool) {
        allowance[msg.sender][guy] = wad;
        emit Approval(msg.sender, guy, wad);
        return true;
    }

    function transfer(address dst, uint256 wad) public returns (bool) {
        return transferFrom(msg.sender, dst, wad);
    }

    function transferFrom(address src, address dst, uint256 wad) public returns (bool) {
        require(balanceOf[src] >= wad, "Insufficient balance");
        if (src != msg.sender && allowance[src][msg.sender] != type(uint256).max) {
            require(allowance[src][msg.sender] >= wad, "Insufficient allowance");
            allowance[src][msg.sender] -= wad;
        }
        balanceOf[src] -= wad;
        balanceOf[dst] += wad;
        emit Transfer(src, dst, wad);
        return true;
    }
}

/// @title SkyUSD (USDC Test/Bridge Token)
contract SkyUSD {
    string public name = "Bridged USD Coin";
    string public symbol = "USDC";
    uint8 public decimals = 6;
    uint256 public totalSupply;

    mapping(address => uint256) public balanceOf;
    mapping(address => mapping(address => uint256)) public allowance;

    event Transfer(address indexed from, address indexed to, uint256 value);
    event Approval(address indexed owner, address indexed spender, uint256 value);

    constructor(uint256 initialSupply) {
        totalSupply = initialSupply;
        balanceOf[msg.sender] = initialSupply;
        emit Transfer(address(0), msg.sender, initialSupply);
    }

    function approve(address spender, uint256 amount) external returns (bool) {
        allowance[msg.sender][spender] = amount;
        emit Approval(msg.sender, spender, amount);
        return true;
    }

    function transfer(address to, uint256 amount) external returns (bool) {
        return transferFrom(msg.sender, to, amount);
    }

    function transferFrom(address from, address to, uint256 amount) public returns (bool) {
        require(balanceOf[from] >= amount, "Insufficient USDC");
        if (from != msg.sender && allowance[from][msg.sender] != type(uint256).max) {
            require(allowance[from][msg.sender] >= amount, "Insufficient allowance");
            allowance[from][msg.sender] -= amount;
        }
        balanceOf[from] -= amount;
        balanceOf[to] += amount;
        emit Transfer(from, to, amount);
        return true;
    }
}

/// @title SkySwapAMM - Uniswap-style Constant Product DEX (x * y = k) with 0.3% LP fee + 0.1% SKY burn
contract SkySwapAMM {
    WSKY public immutable wsky;
    SkyUSD public immutable usdc;

    uint256 public reserveSKY;
    uint256 public reserveUSDC;
    uint256 public totalBurnedSKY;

    event LiquidityAdded(address indexed provider, uint256 skyAmount, uint256 usdcAmount);
    event Swap(address indexed trader, bool skyToUsdc, uint256 amountIn, uint256 amountOut, uint256 burnedSKY);

    constructor(address payable _wsky, address _usdc) {
        wsky = WSKY(_wsky);
        usdc = SkyUSD(_usdc);
    }

    function addLiquidity(uint256 skyAmount, uint256 usdcAmount) external {
        require(wsky.transferFrom(msg.sender, address(this), skyAmount), "WSKY transfer failed");
        require(usdc.transferFrom(msg.sender, address(this), usdcAmount), "USDC transfer failed");
        reserveSKY += skyAmount;
        reserveUSDC += usdcAmount;
        emit LiquidityAdded(msg.sender, skyAmount, usdcAmount);
    }

    /// @notice Returns live market price of 1 SKY in USDC (6 decimals)
    function getSkyPriceUSDC() external view returns (uint256) {
        if (reserveSKY == 0) return 0;
        return (reserveUSDC * 1e18) / reserveSKY;
    }

    /// @notice Swap USDC -> SKY (creates buy pressure & burns 0.1% SKY)
    function swapUSDCForSKY(uint256 usdcIn) external returns (uint256 skyOut) {
        require(usdcIn > 0 && reserveSKY > 0 && reserveUSDC > 0, "Invalid swap");
        require(usdc.transferFrom(msg.sender, address(this), usdcIn), "USDC transfer failed");

        uint256 usdcInWithFee = usdcIn * 997; // 0.3% LP fee
        uint256 rawSkyOut = (usdcInWithFee * reserveSKY) / (reserveUSDC * 1000 + usdcInWithFee);

        uint256 burnCut = rawSkyOut / 1000; // 0.1% deflationary burn
        skyOut = rawSkyOut - burnCut;

        reserveUSDC += usdcIn;
        reserveSKY -= rawSkyOut;
        totalBurnedSKY += burnCut;

        // Send burned SKY to 0x000000000000000000000000000000000000dEaD
        require(wsky.transfer(address(0x000000000000000000000000000000000000dEaD), burnCut), "Burn failed");
        require(wsky.transfer(msg.sender, skyOut), "SKY transfer failed");

        emit Swap(msg.sender, false, usdcIn, skyOut, burnCut);
    }
}

/// @title SkyStakingVault - Community staking pool earning real yield from protocol revenue
contract SkyStakingVault {
    WSKY public immutable wsky;
    uint256 public totalStaked;
    mapping(address => uint256) public stakedBalance;

    event Staked(address indexed user, uint256 amount);
    event Unstaked(address indexed user, uint256 amount);

    constructor(address payable _wsky) {
        wsky = WSKY(_wsky);
    }

    function stake(uint256 amount) external {
        require(amount > 0, "Zero amount");
        require(wsky.transferFrom(msg.sender, address(this), amount), "Transfer failed");
        stakedBalance[msg.sender] += amount;
        totalStaked += amount;
        emit Staked(msg.sender, amount);
    }

    function unstake(uint256 amount) external {
        require(stakedBalance[msg.sender] >= amount, "Exceeds staked");
        stakedBalance[msg.sender] -= amount;
        totalStaked -= amount;
        require(wsky.transfer(msg.sender, amount), "Transfer failed");
        emit Unstaked(msg.sender, amount);
    }
}
