const crypto = require('crypto');

// 模拟 Go 语言中的测试用例
const tests = [
    {
        name: "RSAEncrypt success",
        data: "admin123", // 对应 Go 语言中的 []byte("admin123")
        // 这是 Go 语言测试脚本中 Base64 编码后的公钥
        publicKeyBase64: `LS0tLS1CRUdJTiBQVUJMSUMgS0VZLS0tLS0KTUlJQklqQU5CZ2txaGtpRzl3MEJBUUVGQUFPQ0FROEFNSUlCQ2dLQ0FRRUF6UnN2MjVVNEJqOURqODc4b0JFTwpYREh6ei9LbDFUMTZ3b2hoVzBLejY0VkdOMXFBb09ZTW1Zc2VEd2RHM3R4VXl5Qk5vSVppcTMvOGFLSXhhdDc2Ck9hYlRucXViRlkrODNvRWh4dWUxbWNaRUpTSFlJbkh3UlRjWDhLd3ZXZFZCQWRpUmZYRUR6Zm5TSEV5TUJPM1YKYjNCd2I5dG4vT3BmN2FSbFlxRm1IT1JkYklsRkhmVlpEM0laeTYrUG9tbnFzRVYrUUtkSEdQcWppMVZ2RDhCQgpSb3RnalByQnVkTm1YVnZTMFNoSUZOU3d2aHVSRkM4Y3NZRk1YT0ZQYldVUDU1dlg4MThxM211djBGYU85Um0yCllmZC8ra3JwM3ZFWEVSQURXY2tnTXZrMWJ3OUJaOG5ZbmRHTW5JbldqSjljQlVEK3dZV0NlOGR5bmlDakxBM1YKc1FJREFRQUIKLS0tLS1FTkQgUFVCTElDIEtFWS0tLS0tCg==`
    },
];

console.log("开始运行 RSA 加密测试...");

tests.forEach((testCase, index) => {
    console.log(`
--- 测试用例 ${index + 1}: ${testCase.name} ---`);

    try {
        // 1. Base64 解码公钥
        const publicKeyPem = Buffer.from(testCase.publicKeyBase64, 'base64').toString('utf8');


        // 2. 准备要加密的数据
        const dataToEncrypt = Buffer.from(testCase.data, 'utf8');

        // 3. 使用 RSA-OAEP 填充进行加密 (对应 Go 语言的 rsa.EncryptOAEP)
        const encryptedBuffer = crypto.publicEncrypt(
            {
                key: publicKeyPem,
                padding: crypto.constants.RSA_PKCS1_OAEP_PADDING,
                oaepHash: 'sha256', // 对应 Go 语言的 sha256.New()
            },
            dataToEncrypt
        );

        // 4. 返回 Base64 编码的密文 (对应 Go 语言的 base64.StdEncoding.EncodeToString)
        const encryptedBase64 = encryptedBuffer.toString('base64');

        console.log("原始数据:", testCase.data);
        console.log("加密并 Base64 编码后的结果 (Node.js):", encryptedBase64);

        // 这里可以添加断言来比较结果，如果 Go 语言的测试结果已知的话
        // 例如: assert.strictEqual(encryptedBase64, expectedGoResult);

    } catch (error) {
        console.error("测试失败:", error.message);
    }
});

console.log("RSA 加密测试运行完毕。");