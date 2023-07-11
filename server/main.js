import express from 'express';

const app = express();
app.listen(5001, () => console.log('Server started!'));

app.use('*', async (req, res) => {
    const text = await (await fetch('https://truyenyy.vip' + req.baseUrl)).text();
    res.send(text);
});
